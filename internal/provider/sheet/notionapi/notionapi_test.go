package notionapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

func TestListTablesPaginatesAndFiltersUnavailableTables(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("start_cursor") == "next" {
			_, _ = fmt.Fprint(writer, `{
                    "has_more": false,
                    "results": [
                        {"id":"feb","child_database":{"title":"Feb 2026"}},
                        {"id":"trash","in_trash":true,"child_database":{"title":"Trash"}}
					]
				}`)

			return
		}

		_, _ = fmt.Fprint(writer, `{
                "has_more": true,
                "next_cursor": "next",
                "results": [
                    {"id":"jan","child_database":{"title":"Jan 2026"}},
                    {"id":"archived","archived":true,"child_database":{"title":"Archived"}}
                ]
            }`)
	}))
	t.Cleanup(server.Close)

	client := &Client{
		client: resty.New().SetBaseURL(server.URL),
		conns:  map[string]conn{"sync-profile": {accessToken: "token", pageID: "page"}},
	}

	tables, err := client.ListTables(t.Context(), "sync-profile")
	if err != nil {
		t.Fatalf("ListTables() error = %v", err)
	}

	want := []entity.Table{{ID: "jan", Title: "Jan 2026"}, {ID: "feb", Title: "Feb 2026"}}
	if len(tables) != len(want) {
		t.Fatalf("tables = %#v", tables)
	}
	for index := range want {
		if tables[index] != want[index] {
			t.Fatalf("table[%d] = %#v, want %#v", index, tables[index], want[index])
		}
	}
	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want 2", requests.Load())
	}
}

func TestMapPageToTransaction(t *testing.T) {
	t.Parallel()

	amount := 42.5
	page := listTransactionsRespPage{
		Properties: listTransactionsRespPageProperties{
			Name: listTransactionsRespTitleProperty{Title: []listTransactionsRespRichTextItem{{PlainText: "Store"}}},
			Category: listTransactionsRespSelectProperty{
				Select: &listTransactionsRespSelectOption{Name: "Shopping"},
			},
			Amount: listTransactionsRespNumberProperty{Number: &amount},
			PaymentMethod: listTransactionsRespSelectProperty{
				Select: &listTransactionsRespSelectOption{Name: string(entity.PaymentMethodCreditCard)},
			},
			Date: listTransactionsRespDateProperty{
				Date: &listTransactionsRespDateValue{Start: "2026-08-09T12:00:00Z"},
			},
		},
	}

	transaction, err := (&Client{}).mapPageToTransaction(page)
	if err != nil {
		t.Fatalf("mapPageToTransaction() error = %v", err)
	}
	if transaction.Name != "Store" || transaction.Category != "Shopping" ||
		transaction.Amount != amount || transaction.PaymentMethod != entity.PaymentMethodCreditCard ||
		!transaction.Date.Equal(time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("transaction = %#v", transaction)
	}

	page.Properties.Category.Select = nil
	transaction, err = (&Client{}).mapPageToTransaction(page)
	if err != nil {
		t.Fatalf("mapPageToTransaction() without category error = %v", err)
	}
	if transaction.Category != "" {
		t.Fatalf("missing category = %q, want empty", transaction.Category)
	}
}

func TestCreateTransactionsTableRequestUsesConnectionCategories(t *testing.T) {
	t.Parallel()

	client := &Client{}
	first := client.getRequestData(conn{
		pageID: "page",
		colorsByCategory: map[entity.Category]entity.Color{
			"Food":   entity.Red,
			"Others": entity.Gray,
		},
	}, "Jan 2026")
	second := client.getRequestData(conn{
		pageID: "page",
		colorsByCategory: map[entity.Category]entity.Color{
			"Transporte": entity.LightGray,
			"Outros":     entity.Purple,
		},
	}, "Jan 2026")

	assertCategoryOptions := func(
		t *testing.T,
		options []createTransactionTableReqSelectOption,
		want map[string]entity.Color,
	) {
		t.Helper()

		got := make(map[string]entity.Color, len(options))
		for _, option := range options {
			got[option.Name] = option.Color
		}
		if len(got) != len(want) {
			t.Fatalf("category options = %#v, want %#v", got, want)
		}
		for category, color := range want {
			if got[category] != color {
				t.Fatalf("category options = %#v, want %#v", got, want)
			}
		}
	}

	assertCategoryOptions(t, first.Properties.Category.Select.Options, map[string]entity.Color{
		"Food":   entity.Red,
		"Others": entity.Gray,
	})
	assertCategoryOptions(t, second.Properties.Category.Select.Options, map[string]entity.Color{
		"Transporte": entity.LightGray,
		"Outros":     entity.Purple,
	})
}

func TestInsertTransactionPersistsFallbackCategory(t *testing.T) {
	t.Parallel()

	for _, fallback := range []entity.Category{"Others", "Outros"} {
		t.Run(string(fallback), func(t *testing.T) {
			t.Parallel()

			var requestData insertTransactionReq
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if err := json.NewDecoder(request.Body).Decode(&requestData); err != nil {
					t.Errorf("decode request: %v", err)
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(writer, `{}`)
			}))
			t.Cleanup(server.Close)

			client := &Client{
				client: resty.New().SetBaseURL(server.URL),
				conns:  map[string]conn{"sync-profile": {accessToken: "token"}},
			}
			err := client.InsertTransaction(t.Context(), "sync-profile", "table", entity.Transaction{
				Name:     "Store",
				Category: fallback,
				Amount:   10,
				Date:     time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("InsertTransaction() error = %v", err)
			}
			if requestData.Properties.Category.Select.Name != string(fallback) {
				t.Fatalf(
					"category = %q, want %q",
					requestData.Properties.Category.Select.Name,
					fallback,
				)
			}
		})
	}
}
