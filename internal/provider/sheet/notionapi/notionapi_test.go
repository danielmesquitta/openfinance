package notionapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

func TestListTablesPaginatesAndFiltersUnavailableTables(t *testing.T) {
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

	client := testClient(server.URL)
	tables, err := client.ListTables(t.Context(), "connection")
	if err != nil {
		t.Fatalf("ListTables() error = %v", err)
	}

	want := []sheet.Table{{ID: "jan", Title: "Jan 2026"}, {ID: "feb", Title: "Feb 2026"}}
	if !reflect.DeepEqual(tables, want) {
		t.Fatalf("tables = %#v, want %#v", tables, want)
	}
	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want 2", requests.Load())
	}
}

func TestCreateTableTranslatesGenericDefinition(t *testing.T) {
	var requestData createTableReq
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/databases" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("authorization = %q", request.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(request.Body).Decode(&requestData); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{"id":"table","title":[{"plain_text":"Jan 2026"}]}`)
	}))
	t.Cleanup(server.Close)

	table, err := testClient(server.URL).CreateTable(
		t.Context(),
		"connection",
		"Jan 2026",
		nil,
		sheet.WithIcon("ignored"),
		sheet.WithIcon("💸"),
		sheet.WithColumns(
			sheet.NewTitleColumn("Name"),
			sheet.NewSelectColumn("Category", sheet.WithSelectOptions(
				sheet.NewSelectOption("Food, drinks", sheet.WithColor("red")),
			)),
			sheet.NewNumberColumn("Amount", sheet.WithCurrency("BRL")),
			sheet.NewTextColumn("Notes"),
			sheet.NewDateColumn("Date"),
		),
	)
	if err != nil {
		t.Fatalf("CreateTable() error = %v", err)
	}
	if table != (sheet.Table{ID: "table", Title: "Jan 2026"}) {
		t.Fatalf("table = %#v", table)
	}
	if requestData.Parent.PageID != "page" || requestData.Icon == nil || requestData.Icon.Emoji != "💸" ||
		len(requestData.Title) != 1 || requestData.Title[0].Text.Content != "Jan 2026" {
		t.Fatalf("request metadata = %#v", requestData)
	}
	if requestData.Properties["Name"].Title == nil || requestData.Properties["Notes"].RichText == nil ||
		requestData.Properties["Date"].Date == nil {
		t.Fatalf("request properties = %#v", requestData.Properties)
	}
	if number := requestData.Properties["Amount"].Number; number == nil || number.Format != "real" {
		t.Fatalf("amount property = %#v", requestData.Properties["Amount"])
	}
	options := requestData.Properties["Category"].Select.Options
	if len(options) != 1 || options[0].Name != "Food drinks" || options[0].Color != "red" {
		t.Fatalf("category options = %#v", options)
	}
}

func TestCreateTableRequestSupportsEmptyOptionalMetadata(t *testing.T) {
	requestData, err := createTableRequest("page", "Table", sheet.CreateTableOptions{
		Columns: []sheet.Column{
			sheet.NewNumberColumn("Amount"),
			sheet.NewSelectColumn("Category"),
		},
	})
	if err != nil {
		t.Fatalf("createTableRequest() error = %v", err)
	}
	if requestData.Icon != nil {
		t.Fatalf("icon = %#v, want nil", requestData.Icon)
	}
	if number := requestData.Properties["Amount"].Number; number == nil || number.Format != "" {
		t.Fatalf("amount property = %#v", requestData.Properties["Amount"])
	}
	if options := requestData.Properties["Category"].Select.Options; len(options) != 0 {
		t.Fatalf("category options = %#v, want empty", options)
	}
}

func TestCreateTableRejectsInvalidOptionsBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	t.Cleanup(server.Close)
	client := testClient(server.URL)

	tests := []sheet.Column{
		sheet.NewNumberColumn("Amount", sheet.WithCurrency("USD")),
		{},
		sheet.NewTitleColumn(""),
		sheet.NewSelectColumn("Category", sheet.WithSelectOptions(sheet.NewSelectOption(""))),
	}
	for _, column := range tests {
		_, err := client.CreateTable(
			t.Context(),
			"connection",
			"Table",
			sheet.WithColumns(column),
		)
		if err == nil {
			t.Fatalf("CreateTable(%#v) error = nil", column)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("requests = %d, want 0", requests.Load())
	}
}

func TestInsertRowTranslatesEveryCellType(t *testing.T) {
	var requestData insertRowReq
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/pages" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&requestData); err != nil {
			t.Errorf("decode request: %v", err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{}`)
	}))
	t.Cleanup(server.Close)

	date := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	err := testClient(server.URL).InsertRow(t.Context(), "connection", "table", sheet.Row{
		"Name":     sheet.TitleCell("Store"),
		"Notes":    sheet.TextCell("note"),
		"Amount":   sheet.NumberCell(42.5),
		"Category": sheet.SelectCell("Food, drinks"),
		"Date":     sheet.DateCell(date),
	})
	if err != nil {
		t.Fatalf("InsertRow() error = %v", err)
	}

	if requestData.Parent.DatabaseID != "table" ||
		requestData.Properties["Name"].Title[0].Text.Content != "Store" ||
		requestData.Properties["Notes"].RichText[0].Text.Content != "note" ||
		*requestData.Properties["Amount"].Number != 42.5 ||
		requestData.Properties["Category"].Select.Name != "Food drinks" ||
		requestData.Properties["Date"].Date.Start != date.Format(time.RFC3339) {
		t.Fatalf("request = %#v", requestData)
	}
}

func TestInsertRowRejectsNilCellBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	t.Cleanup(server.Close)

	err := testClient(server.URL).InsertRow(t.Context(), "connection", "table", sheet.Row{"Name": nil})
	if err == nil {
		t.Fatal("InsertRow() error = nil")
	}
	if requests.Load() != 0 {
		t.Fatalf("requests = %d, want 0", requests.Load())
	}
}

func TestListRowsPaginatesMapsPropertiesAndSkipsMalformedDates(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		var body listRowsReq
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.PageSize != maxPageSize || len(body.Sorts) != 1 ||
			body.Sorts[0].Timestamp != "created_time" || body.Sorts[0].Direction != "descending" {
			t.Errorf("query request = %#v", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		if body.StartCursor == "next" {
			_, _ = fmt.Fprint(writer, `{
				"has_more":false,
				"results":[{"properties":{
					"Name":{"type":"title","title":[]},
					"Notes":{"type":"rich_text","rich_text":[]},
					"Category":{"type":"select","select":null},
					"Amount":{"type":"number","number":null},
					"Date":{"type":"date","date":null}
				}}]
			}`)

			return
		}

		_, _ = fmt.Fprint(writer, `{
			"has_more":true,
			"next_cursor":"next",
			"results":[
				{"properties":{
					"Name":{"type":"title","title":[{"plain_text":"Store"},{"plain_text":"ignored"}]},
					"Notes":{"type":"rich_text","rich_text":[{"plain_text":"first"},{"plain_text":"ignored"}]},
					"Category":{"type":"select","select":{"name":"Food"}},
					"Amount":{"type":"number","number":42.5},
					"Date":{"type":"date","date":{"start":"2026-08-09T12:00:00Z"}},
					"Ignored":{"type":"formula","formula":{"type":"string","string":"x"}}
				}},
				{"properties":{"Date":{"type":"date","date":{"start":"not-a-date"}}}}
			]
		}`)
	}))
	t.Cleanup(server.Close)

	rows, err := testClient(server.URL).ListRows(t.Context(), "connection", "table")
	if err != nil {
		t.Fatalf("ListRows() error = %v", err)
	}
	if requests.Load() != 2 || len(rows) != 2 {
		t.Fatalf("requests = %d, rows = %#v", requests.Load(), rows)
	}

	wantDate := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	if rows[0]["Name"] != sheet.TitleCell("Store") ||
		rows[0]["Notes"] != sheet.TextCell("first") ||
		rows[0]["Category"] != sheet.SelectCell("Food") ||
		rows[0]["Amount"] != sheet.NumberCell(42.5) ||
		time.Time(rows[0]["Date"].(sheet.DateCell)) != wantDate {
		t.Fatalf("first row = %#v", rows[0])
	}
	if _, exists := rows[0]["Ignored"]; exists {
		t.Fatalf("unsupported property was mapped: %#v", rows[0])
	}
	if rows[1]["Name"] != sheet.TitleCell("") || rows[1]["Notes"] != sheet.TextCell("") || len(rows[1]) != 2 {
		t.Fatalf("empty row = %#v", rows[1])
	}
}

func TestOperationsRejectMissingConnection(t *testing.T) {
	client := &Client{client: resty.New(), conns: map[string]conn{}}
	if _, err := client.CreateTable(t.Context(), "missing", "Table"); err == nil {
		t.Fatal("CreateTable() error = nil")
	}
	if err := client.InsertRow(t.Context(), "missing", "table", nil); err == nil {
		t.Fatal("InsertRow() error = nil")
	}
	if _, err := client.ListTables(t.Context(), "missing"); err == nil {
		t.Fatal("ListTables() error = nil")
	}
	if _, err := client.ListRows(t.Context(), "missing", "table"); err == nil {
		t.Fatal("ListRows() error = nil")
	}
}

func testClient(baseURL string) *Client {
	return &Client{
		client: resty.New().SetBaseURL(baseURL),
		conns: map[string]conn{
			"connection": {accessToken: "token", pageID: "page"},
		},
	}
}
