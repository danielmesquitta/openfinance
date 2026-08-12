package pluggyapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

func TestListTransactionsByIngestProfileIDPaginatesAndMaps(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		if request.URL.Path != "/transactions" {
			http.NotFound(writer, request)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Query().Get("page") {
		case "1":
			_, _ = fmt.Fprint(writer, `{
                "totalPages": 2,
                "page": 1,
                "results": [{
                    "description": "Store",
                    "amount": -10.5,
                    "date": "2026-08-09T12:00:00Z",
                    "type": "DEBIT",
                    "creditCardMetadata": {"cardNumber": "1234"}
                }]
            }`)
		case "2":
			_, _ = fmt.Fprint(writer, `{
                "totalPages": 2,
                "page": 2,
                "results": [{
                    "amount": -20,
                    "date": "2026-08-10T12:00:00Z",
                    "type": "DEBIT",
                    "paymentData": {
                        "paymentMethod": "PIX",
                        "receiver": {"documentNumber": {"value": "12345678000195"}}
                    }
                }]
            }`)
		default:
			http.Error(writer, "unexpected page", http.StatusBadRequest)
		}
	}))
	t.Cleanup(server.Close)

	client := &Client{
		client:                  resty.New().SetBaseURL(server.URL),
		maxConcurrentOperations: 1,
		conns: map[string]conn{
			"ingest-profile": {accessToken: "token", accountIDs: []string{"account"}},
		},
		accountSlots: make(chan struct{}, 1),
	}

	transactions, err := client.ListTransactionsByIngestProfileID(
		t.Context(),
		"ingest-profile",
		time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ListTransactionsByIngestProfileID() error = %v", err)
	}

	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want 2", requests.Load())
	}
	if len(transactions) != 2 {
		t.Fatalf("transactions = %#v", transactions)
	}
	if transactions[0].Name != "Store" ||
		transactions[0].PaymentMethod != entity.PaymentMethodCreditCard ||
		transactions[0].Amount != 10.5 {
		t.Fatalf("credit card transaction = %#v", transactions[0])
	}
	if transactions[1].Name != "12.345.678/0001-95" ||
		transactions[1].PaymentMethod != entity.PaymentMethodPix ||
		transactions[1].Amount != 20 {
		t.Fatalf("bank transaction = %#v", transactions[1])
	}
}
