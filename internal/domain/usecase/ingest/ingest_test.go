package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi/mockcompanyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt/mockgpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/mockopenfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/mocksheet"
)

type insertedTransaction struct {
	tableID     string
	transaction entity.Transaction
}

const testMaxConcurrentOperations = 4

type categorizationInput struct {
	TransactionNames []string                   `json:"transaction_names"`
	Categories       []entity.Category          `json:"categories"`
	Mappings         map[string]entity.Category `json:"examples"`
	Fallback         entity.Category            `json:"fallback"`
}

func applyChatCompletionOptions(options []gpt.ChatCompletionOption) gpt.ChatCompletionOptions {
	completionOptions := gpt.ChatCompletionOptions{}
	for _, option := range options {
		if option != nil {
			option(&completionOptions)
		}
	}

	return completionOptions
}

func setMaximum(maximum *atomic.Int32, value int32) {
	for {
		current := maximum.Load()
		if value <= current || maximum.CompareAndSwap(current, value) {
			return
		}
	}
}

func testIngestProfileSettings(ingestProfileID string) entity.IngestProfileSettings {
	return entity.IngestProfileSettings{
		ID:         ingestProfileID,
		Categories: []entity.Category{"Food", entity.DefaultFallbackCategory},
		ColorsByCategory: map[entity.Category]entity.Color{
			"Food":                         entity.Red,
			entity.DefaultFallbackCategory: entity.Gray,
		},
		Mappings: map[string]entity.Category{"Market": "Food"},
		Fallback: entity.DefaultFallbackCategory,
	}
}

func testSettings(ingestProfileIDs ...string) entity.IngestSettings {
	settings := entity.IngestSettings{IngestProfiles: make([]entity.IngestProfileSettings, 0, len(ingestProfileIDs))}
	for _, ingestProfileID := range ingestProfileIDs {
		settings.IngestProfiles = append(settings.IngestProfiles, testIngestProfileSettings(ingestProfileID))
	}

	return settings
}

func noCompanyLookup(t *testing.T) *mockcompanyapi.MockCompanyAPI {
	t.Helper()

	return mockcompanyapi.NewMockCompanyAPI(t)
}

func TestIngestInputValidate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []IngestInput{
		{},
		{StartDate: now},
		{StartDate: now, EndDate: now.Add(-time.Second)},
	}

	for _, input := range tests {
		if err := input.Validate(); err == nil {
			t.Fatalf("Validate(%#v) error = nil", input)
		}
	}

	if err := (IngestInput{StartDate: now, EndDate: now}).Validate(); err != nil {
		t.Fatalf("valid input error = %v", err)
	}
}

func TestCategorizeTransactionsRejectsInvalidCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{name: "empty"},
		{name: "malformed", content: "not JSON"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			categorizer := mockgpt.NewMockGPT(t)
			categorizer.EXPECT().
				CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
				Return(test.content, nil).
				Once()

			ingestUseCase := &Ingest{
				settings:    testSettings(),
				gptProvider: categorizer,
			}
			err := ingestUseCase.categorizeTransactions(
				t.Context(),
				testIngestProfileSettings("ingest-profile"),
				[]entity.Transaction{{Name: "Store"}},
			)
			if err == nil {
				t.Fatal("categorizeTransactions() error = nil")
			}
		})
	}
}

func TestCategorizeTransactionsUsesFallbackForMissingAndInvalidCategories(t *testing.T) {
	t.Parallel()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		Return(`{"Invalid":"not-configured"}`, nil).
		Once()

	transactions := []entity.Transaction{{Name: "Missing"}, {Name: "Invalid"}}
	ingestUseCase := &Ingest{
		settings:    testSettings(),
		gptProvider: categorizer,
	}
	if err := ingestUseCase.categorizeTransactions(
		t.Context(),
		testIngestProfileSettings("ingest-profile"),
		transactions,
	); err != nil {
		t.Fatalf("categorizeTransactions() error = %v", err)
	}

	for _, transaction := range transactions {
		if transaction.Category != entity.DefaultFallbackCategory {
			t.Fatalf("transaction = %#v", transaction)
		}
	}
}

func TestIngestUsesIngestProfileSpecificCategorizationSettings(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	settings := entity.IngestSettings{IngestProfiles: []entity.IngestProfileSettings{
		{
			ID:               "first-ingest-profile",
			Categories:       []entity.Category{"Alpha", "Others"},
			ColorsByCategory: map[entity.Category]entity.Color{"Alpha": entity.Red, "Others": entity.Gray},
			Mappings:         map[string]entity.Category{"First example": "Alpha"},
			Fallback:         "Others",
		},
		{
			ID:               "second-ingest-profile",
			Categories:       []entity.Category{"Beta", "Outros"},
			ColorsByCategory: map[entity.Category]entity.Color{"Beta": entity.Blue, "Outros": entity.Purple},
			Mappings:         map[string]entity.Category{"Second example": "Beta"},
			Fallback:         "Outros",
		},
	}}

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, mock.Anything, date, date).
		RunAndReturn(func(_ context.Context, ingestProfileID string, _, _ time.Time) ([]entity.Transaction, error) {
			return []entity.Transaction{{Name: ingestProfileID + " transaction", Amount: 10, Date: date}}, nil
		}).
		Twice()

	inputs := make(map[string]categorizationInput, 2)
	var inputsMutex sync.Mutex
	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			message string,
			_ ...gpt.ChatCompletionOption,
		) (string, error) {
			var input categorizationInput
			if err := json.Unmarshal([]byte(message), &input); err != nil {
				return "", fmt.Errorf("decode categorization input: %w", err)
			}
			if len(input.TransactionNames) != 1 {
				return "", fmt.Errorf("transaction names = %#v", input.TransactionNames)
			}

			name := input.TransactionNames[0]
			inputsMutex.Lock()
			inputs[name] = input
			inputsMutex.Unlock()

			if name == "first-ingest-profile transaction" {
				return `{"first-ingest-profile transaction":"Beta"}`, nil
			}

			return `{"second-ingest-profile transaction":"Alpha"}`, nil
		}).
		Twice()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, mock.Anything).Return(nil, nil).Twice()
	store.EXPECT().
		CreateTable(mock.Anything, mock.Anything, "Jan 2026", mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			ingestProfileID, title string,
			_ ...sheet.CreateTableOption,
		) (sheet.Table, error) {
			return sheet.Table{ID: ingestProfileID + "-table", Title: title}, nil
		}).
		Twice()

	insertedCategories := make(map[string]entity.Category, 2)
	var insertedMutex sync.Mutex
	store.EXPECT().
		InsertRow(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, ingestProfileID, _ string, row sheet.Row) {
			transaction, err := rowToTransaction(row, entity.LanguageEnglish)
			if err != nil {
				t.Errorf("rowToTransaction() error = %v", err)

				return
			}
			insertedMutex.Lock()
			insertedCategories[ingestProfileID] = transaction.Category
			insertedMutex.Unlock()
		}).
		Return(nil).
		Twice()

	err := NewIngest(
		testMaxConcurrentOperations,
		settings,
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(context.Background(), IngestInput{StartDate: date, EndDate: date})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	firstInput := inputs["first-ingest-profile transaction"]
	if len(firstInput.Categories) != 2 || firstInput.Categories[0] != "Alpha" ||
		firstInput.Mappings["First example"] != "Alpha" || firstInput.Fallback != "Others" {
		t.Fatalf("first categorization input = %#v", firstInput)
	}
	secondInput := inputs["second-ingest-profile transaction"]
	if len(secondInput.Categories) != 2 || secondInput.Categories[0] != "Beta" ||
		secondInput.Mappings["Second example"] != "Beta" || secondInput.Fallback != "Outros" {
		t.Fatalf("second categorization input = %#v", secondInput)
	}
	if insertedCategories["first-ingest-profile"] != "Others" ||
		insertedCategories["second-ingest-profile"] != "Outros" {
		t.Fatalf("inserted categories = %#v", insertedCategories)
	}
}

func TestIngestProcessesEveryMonthAndDeduplicates(t *testing.T) {
	t.Parallel()

	janDate := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	febDate := time.Date(2026, time.February, 10, 12, 0, 0, 0, time.UTC)
	existing := entity.Transaction{Name: "Market", Amount: 10, Date: janDate}
	newJanuary := entity.Transaction{Name: "Cafe", Amount: 20, Date: janDate}
	newFebruary := entity.Transaction{Name: "Unknown", Amount: 30, Date: febDate}

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", janDate, febDate).
		Return([]entity.Transaction{existing, existing, newJanuary, newFebruary}, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			message string,
			options ...gpt.ChatCompletionOption,
		) (string, error) {
			var input categorizationInput
			if err := json.Unmarshal([]byte(message), &input); err != nil {
				t.Fatalf("categorization input = %q: %v", message, err)
			}
			if len(input.TransactionNames) != 3 {
				t.Fatalf("unique names = %#v", input.TransactionNames)
			}
			if len(input.Categories) != 2 || input.Categories[0] != "Food" {
				t.Fatalf("categories = %#v", input.Categories)
			}
			if input.Mappings["Market"] != "Food" || input.Fallback != entity.DefaultFallbackCategory {
				t.Fatalf("categorization input = %#v", input)
			}

			completionOptions := applyChatCompletionOptions(options)
			if completionOptions.SystemMessage != categorizationSystemMessage {
				t.Fatalf("system message = %q", completionOptions.SystemMessage)
			}
			if !completionOptions.JSONResponse {
				t.Fatal("JSON response option is disabled")
			}

			return `{"Market":"Food","Cafe":"Food","Unknown":"not-configured"}`, nil
		}).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().
		ListTables(mock.Anything, "ingest-profile").
		Return([]sheet.Table{{ID: "jan", Title: "Jan 2026"}}, nil).
		Once()
	store.EXPECT().
		ListRows(mock.Anything, "ingest-profile", "jan").
		Return([]sheet.Row{transactionToRow(existing, entity.LanguageEnglish)}, nil).
		Once()
	store.EXPECT().
		CreateTable(
			mock.Anything,
			"ingest-profile",
			"Feb 2026",
			mock.Anything,
		).
		Return(sheet.Table{ID: "created-Feb 2026", Title: "Feb 2026"}, nil).
		Once()

	var insertedMutex sync.Mutex
	inserted := make([]insertedTransaction, 0, 2)
	store.EXPECT().
		InsertRow(mock.Anything, "ingest-profile", mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ string, tableID string, row sheet.Row) {
			transaction, err := rowToTransaction(row, entity.LanguageEnglish)
			if err != nil {
				t.Errorf("rowToTransaction() error = %v", err)

				return
			}
			insertedMutex.Lock()
			inserted = append(inserted, insertedTransaction{tableID: tableID, transaction: transaction})
			insertedMutex.Unlock()
		}).
		Return(nil).
		Twice()

	ingestUseCase := NewIngest(
		testMaxConcurrentOperations,
		testSettings("ingest-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	)
	err := ingestUseCase.Execute(context.Background(), IngestInput{
		StartDate: janDate,
		EndDate:   febDate,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(inserted) != 2 {
		t.Fatalf("inserted = %#v, want 2 transactions", inserted)
	}
	if inserted[0].tableID != "jan" || inserted[0].transaction.Name != "Cafe" {
		t.Fatalf("January insert = %#v", inserted[0])
	}
	if inserted[1].tableID != "created-Feb 2026" ||
		inserted[1].transaction.Category != entity.DefaultFallbackCategory {
		t.Fatalf("February insert = %#v", inserted[1])
	}
}

func TestIngestReusesExistingTableLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		configuredLanguage    entity.Language
		existingTitle         string
		existingTableLanguage entity.Language
		paymentMethodColumn   string
		paymentMethodLabel    sheet.SelectCell
	}{
		{
			name:                  "Portuguese profile reuses English table",
			configuredLanguage:    entity.LanguagePortugueseBrazil,
			existingTitle:         "Aug 2026",
			existingTableLanguage: entity.LanguageEnglish,
			paymentMethodColumn:   "Payment Method",
			paymentMethodLabel:    "CREDIT CARD",
		},
		{
			name:                  "English profile reuses Portuguese table",
			configuredLanguage:    entity.LanguageEnglish,
			existingTitle:         "Ago 2026",
			existingTableLanguage: entity.LanguagePortugueseBrazil,
			paymentMethodColumn:   "Forma de pagamento",
			paymentMethodLabel:    "CARTÃO DE CRÉDITO",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			date := time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC)
			source := mockopenfinance.NewMockOpenFinance(t)
			source.EXPECT().
				ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
				Return([]entity.Transaction{{
					Name:          "Store",
					Amount:        10,
					PaymentMethod: entity.PaymentMethodCreditCard,
					Date:          date,
				}}, nil).
				Once()

			categorizer := mockgpt.NewMockGPT(t)
			categorizer.EXPECT().
				CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
				Return(`{"Store":"Food"}`, nil).
				Once()

			store := mocksheet.NewMockSheet(t)
			store.EXPECT().
				ListTables(mock.Anything, "ingest-profile").
				Return([]sheet.Table{{ID: "existing", Title: test.existingTitle}}, nil).
				Once()
			store.EXPECT().
				ListRows(mock.Anything, "ingest-profile", "existing").
				Return(nil, nil).
				Once()
			store.EXPECT().
				InsertRow(
					mock.Anything,
					"ingest-profile",
					"existing",
					mock.MatchedBy(func(row sheet.Row) bool {
						transaction, err := rowToTransaction(row, test.existingTableLanguage)

						return err == nil && transaction.Name == "Store" &&
							transaction.PaymentMethod == entity.PaymentMethodCreditCard &&
							row[test.paymentMethodColumn] == test.paymentMethodLabel
					}),
				).
				Return(nil).
				Once()

			settings := testSettings("ingest-profile")
			settings.IngestProfiles[0].Language = test.configuredLanguage
			if err := NewIngest(
				testMaxConcurrentOperations,
				settings,
				noCompanyLookup(t),
				categorizer,
				store,
				source,
			).Execute(context.Background(), IngestInput{StartDate: date, EndDate: date}); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

func TestIngestCreatesPortugueseTable(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
		Return(nil, nil).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			_, title string,
			options ...sheet.CreateTableOption,
		) (sheet.Table, error) {
			resolved := resolveCreateTableOptions(options)
			if len(resolved.Columns) != 6 || resolved.Columns[0].Name() != "Nome" ||
				resolved.Columns[3].Name() != "Forma de pagamento" {
				t.Fatalf("table options = %#v", resolved)
			}

			return sheet.Table{ID: "january", Title: title}, nil
		}).
		Once()

	settings := testSettings("ingest-profile")
	settings.IngestProfiles[0].Language = entity.LanguagePortugueseBrazil
	if err := NewIngest(
		testMaxConcurrentOperations,
		settings,
		noCompanyLookup(t),
		mockgpt.NewMockGPT(t),
		store,
		source,
	).Execute(context.Background(), IngestInput{StartDate: date, EndDate: date}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestTransactionTableForMonthPrefersConfiguredLanguage(t *testing.T) {
	t.Parallel()

	month := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	tables := map[string]sheet.Table{
		"Aug 2026": {ID: "english", Title: "Aug 2026"},
		"Ago 2026": {ID: "portuguese", Title: "Ago 2026"},
	}

	table, language, exists := transactionTableForMonth(
		tables,
		month,
		entity.LanguagePortugueseBrazil,
	)
	if !exists || table.ID != "portuguese" || language != entity.LanguagePortugueseBrazil {
		t.Fatalf("table = %#v, language = %q, exists = %t", table, language, exists)
	}
}

func TestTransactionTableForMonthUsesConfiguredLanguageForSharedTitle(t *testing.T) {
	t.Parallel()

	month := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tables := map[string]sheet.Table{
		"Jan 2026": {ID: "shared", Title: "Jan 2026"},
	}

	for _, language := range []entity.Language{
		entity.LanguageEnglish,
		entity.LanguagePortugueseBrazil,
	} {
		t.Run(string(language), func(t *testing.T) {
			t.Parallel()

			table, tableLanguage, exists := transactionTableForMonth(tables, month, language)
			if !exists || table.ID != "shared" || tableLanguage != language {
				t.Fatalf(
					"table = %#v, language = %q, exists = %t",
					table,
					tableLanguage,
					exists,
				)
			}
		})
	}
}

func TestGroupTransactionsByMonthUsesStableKeys(t *testing.T) {
	t.Parallel()

	transactions := []entity.Transaction{
		{Date: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{Date: time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)},
		{Date: time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)},
	}

	grouped := groupTransactionsByMonth(transactions)
	if len(grouped[newTransactionMonth(transactions[0].Date)]) != 2 ||
		len(grouped[newTransactionMonth(transactions[2].Date)]) != 1 {
		t.Fatalf("grouped transactions = %#v", grouped)
	}
}

func TestIngestEnrichesUniqueCompany(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	company := mockcompanyapi.NewMockCompanyAPI(t)
	company.EXPECT().
		GetCompanyByID(mock.Anything, "12345678000195").
		Return(entity.Company{TradingName: "Company"}, nil).
		Once()

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
		Return([]entity.Transaction{
			{Name: "12.345.678/0001-95", Amount: 1, Date: date},
			{Name: "12.345.678/0001-95", Amount: 2, Date: date},
		}, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		Return(`{"Company":"Food"}`, nil).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
		Return(sheet.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertRow(
			mock.Anything,
			"ingest-profile",
			"jan",
			mock.MatchedBy(func(row sheet.Row) bool {
				transaction, err := rowToTransaction(row, entity.LanguageEnglish)

				return err == nil && transaction.Name == "Company" && transaction.Category == "Food"
			}),
		).
		Return(nil).
		Twice()

	if err := NewIngest(
		testMaxConcurrentOperations,
		testSettings("ingest-profile"),
		company,
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		IngestInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestIngestCompanyLookupFailureIsNonFatal(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	companyErr := errors.New("company unavailable")
	company := mockcompanyapi.NewMockCompanyAPI(t)
	company.EXPECT().
		GetCompanyByID(mock.Anything, "12345678000195").
		Return(entity.Company{}, companyErr).
		Once()

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
		Return([]entity.Transaction{{Name: "12.345.678/0001-95", Amount: 1, Date: date}}, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		Return(`{"12.345.678/0001-95":"Food"}`, nil).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
		Return(sheet.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertRow(
			mock.Anything,
			"ingest-profile",
			"jan",
			mock.MatchedBy(func(row sheet.Row) bool {
				transaction, err := rowToTransaction(row, entity.LanguageEnglish)

				return err == nil && transaction.Name == "12.345.678/0001-95"
			}),
		).
		Return(nil).
		Once()

	if err := NewIngest(
		testMaxConcurrentOperations,
		testSettings("ingest-profile"),
		company,
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		IngestInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestIngestEmptyRangeSkipsCategorizer(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
		Return(nil, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
		Return(sheet.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()

	if err := NewIngest(
		testMaxConcurrentOperations,
		testSettings("ingest-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		IngestInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestIngestPropagatesSourceErrorAndCancellation(t *testing.T) {
	t.Parallel()

	date := time.Now()
	t.Run("source error", func(t *testing.T) {
		wantErr := errors.New("source failed")
		source := mockopenfinance.NewMockOpenFinance(t)
		source.EXPECT().
			ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
			Return(nil, wantErr).
			Once()

		err := NewIngest(
			testMaxConcurrentOperations,
			testSettings("ingest-profile"),
			noCompanyLookup(t),
			mockgpt.NewMockGPT(t),
			mocksheet.NewMockSheet(t),
			source,
		).Execute(context.Background(), IngestInput{StartDate: date, EndDate: date})
		if !errors.Is(err, wantErr) {
			t.Fatalf("Execute() error = %v, want source error", err)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		canceledContext, cancel := context.WithCancel(context.Background())
		cancel()

		source := mockopenfinance.NewMockOpenFinance(t)
		source.EXPECT().
			ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
			RunAndReturn(func(ctx context.Context, _ string, _, _ time.Time) ([]entity.Transaction, error) {
				return nil, ctx.Err()
			}).
			Once()

		err := NewIngest(
			testMaxConcurrentOperations,
			testSettings("ingest-profile"),
			noCompanyLookup(t),
			mockgpt.NewMockGPT(t),
			mocksheet.NewMockSheet(t),
			source,
		).Execute(canceledContext, IngestInput{StartDate: date, EndDate: date})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Execute() error = %v, want context canceled", err)
		}
	})
}

func TestIngestPropagatesCategorizerAndStoreErrors(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	transaction := entity.Transaction{Name: "Store", Amount: 1, Date: date}
	categorizerErr := errors.New("categorizer failed")
	listTablesErr := errors.New("list tables failed")
	insertErr := errors.New("insert failed")

	tests := []struct {
		name    string
		stage   string
		wantErr error
	}{
		{name: "categorizer", stage: "categorizer", wantErr: categorizerErr},
		{name: "list tables", stage: "list tables", wantErr: listTablesErr},
		{name: "insert", stage: "insert", wantErr: insertErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := mockopenfinance.NewMockOpenFinance(t)
			source.EXPECT().
				ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
				Return([]entity.Transaction{transaction}, nil).
				Once()

			categorizer := mockgpt.NewMockGPT(t)
			store := mocksheet.NewMockSheet(t)
			switch test.stage {
			case "categorizer":
				categorizer.EXPECT().
					CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
					Return("", categorizerErr).
					Once()
			case "list tables":
				categorizer.EXPECT().
					CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
					Return(`{"Store":"Food"}`, nil).
					Once()
				store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, listTablesErr).Once()
			case "insert":
				categorizer.EXPECT().
					CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
					Return(`{"Store":"Food"}`, nil).
					Once()
				store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
				store.EXPECT().
					CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
					Return(sheet.Table{ID: "jan", Title: "Jan 2026"}, nil).
					Once()
				store.EXPECT().
					InsertRow(mock.Anything, "ingest-profile", "jan", mock.Anything).
					Return(insertErr).
					Once()
			}

			err := NewIngest(
				testMaxConcurrentOperations,
				testSettings("ingest-profile"),
				noCompanyLookup(t),
				categorizer,
				store,
				source,
			).Execute(
				context.Background(),
				IngestInput{StartDate: date, EndDate: date},
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Execute() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestIngestBoundsIngestProfileConcurrency(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	ingestProfileIDs := make([]string, 12)
	for index := range ingestProfileIDs {
		ingestProfileIDs[index] = fmt.Sprintf("ingest-profile-%d", index)
	}

	var activeSources atomic.Int32
	var maximumSources atomic.Int32
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, mock.Anything, date, date).
		RunAndReturn(func(ctx context.Context, _ string, _, _ time.Time) ([]entity.Transaction, error) {
			active := activeSources.Add(1)
			defer activeSources.Add(-1)
			setMaximum(&maximumSources, active)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}

			return nil, nil
		}).
		Times(len(ingestProfileIDs))

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, mock.Anything).Return(nil, nil).Times(len(ingestProfileIDs))
	store.EXPECT().
		CreateTable(mock.Anything, mock.Anything, "Jan 2026", mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			ingestProfileID, title string,
			_ ...sheet.CreateTableOption,
		) (sheet.Table, error) {
			return sheet.Table{ID: ingestProfileID + "-jan", Title: title}, nil
		}).
		Times(len(ingestProfileIDs))

	if err := NewIngest(
		testMaxConcurrentOperations,
		testSettings(ingestProfileIDs...),
		noCompanyLookup(t),
		mockgpt.NewMockGPT(t),
		store,
		source,
	).Execute(context.Background(), IngestInput{StartDate: date, EndDate: date}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if maximumSources.Load() > testMaxConcurrentOperations {
		t.Fatalf("source concurrency = %d, limit %d", maximumSources.Load(), testMaxConcurrentOperations)
	}
	if maximumSources.Load() < 2 {
		t.Fatalf("expected concurrent execution, got sources=%d", maximumSources.Load())
	}
}

func TestIngestBoundsInsertConcurrency(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	transactions := make([]entity.Transaction, 12)
	for index := range transactions {
		transactions[index] = entity.Transaction{
			Name:   fmt.Sprintf("transaction-%d", index),
			Amount: float64(index + 1),
			Date:   date,
		}
	}

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByIngestProfileID(mock.Anything, "ingest-profile", date, date).
		Return(transactions, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context,
			message string,
			_ ...gpt.ChatCompletionOption,
		) (string, error) {
			var input categorizationInput
			if err := json.Unmarshal([]byte(message), &input); err != nil {
				return "", fmt.Errorf("decode categorization input: %w", err)
			}

			categories := make(map[string]entity.Category, len(input.TransactionNames))
			for _, name := range input.TransactionNames {
				categories[name] = "Food"
			}

			response, err := json.Marshal(categories)

			return string(response), err
		}).
		Once()

	var activeInserts atomic.Int32
	var maximumInserts atomic.Int32
	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "ingest-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTable(mock.Anything, "ingest-profile", "Jan 2026", mock.Anything).
		Return(sheet.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertRow(mock.Anything, "ingest-profile", "jan", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ string, _ sheet.Row) error {
			active := activeInserts.Add(1)
			defer activeInserts.Add(-1)
			setMaximum(&maximumInserts, active)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Millisecond):
				return nil
			}
		}).
		Times(len(transactions))

	if err := NewIngest(
		testMaxConcurrentOperations,
		testSettings("ingest-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		IngestInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if maximumInserts.Load() > testMaxConcurrentOperations {
		t.Fatalf("insert concurrency = %d, limit %d", maximumInserts.Load(), testMaxConcurrentOperations)
	}
	if maximumInserts.Load() < 2 {
		t.Fatalf("expected concurrent execution, got inserts=%d", maximumInserts.Load())
	}
}
