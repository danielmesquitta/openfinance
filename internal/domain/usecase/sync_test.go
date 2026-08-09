package usecase

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

func testSyncProfileSettings(syncProfileID string) entity.SyncProfileSettings {
	return entity.SyncProfileSettings{
		ID:               syncProfileID,
		Categories:       []entity.Category{"Food", entity.DefaultFallbackCategory},
		ColorsByCategory: map[entity.Category]entity.Color{"Food": entity.Red, entity.DefaultFallbackCategory: entity.Gray},
		Mappings:         map[string]entity.Category{"Market": "Food"},
		Fallback:         entity.DefaultFallbackCategory,
	}
}

func testSettings(syncProfileIDs ...string) entity.SyncSettings {
	settings := entity.SyncSettings{SyncProfiles: make([]entity.SyncProfileSettings, 0, len(syncProfileIDs))}
	for _, syncProfileID := range syncProfileIDs {
		settings.SyncProfiles = append(settings.SyncProfiles, testSyncProfileSettings(syncProfileID))
	}

	return settings
}

func noCompanyLookup(t *testing.T) *mockcompanyapi.MockCompanyAPI {
	t.Helper()

	return mockcompanyapi.NewMockCompanyAPI(t)
}

func TestSyncInputValidate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []SyncInput{
		{},
		{StartDate: now},
		{StartDate: now, EndDate: now.Add(-time.Second)},
	}

	for _, input := range tests {
		if err := input.Validate(); err == nil {
			t.Fatalf("Validate(%#v) error = nil", input)
		}
	}

	if err := (SyncInput{StartDate: now, EndDate: now}).Validate(); err != nil {
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

			syncUseCase := &Sync{
				settings:    testSettings(),
				gptProvider: categorizer,
			}
			err := syncUseCase.categorizeTransactions(
				t.Context(),
				testSyncProfileSettings("sync-profile"),
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
	syncUseCase := &Sync{
		settings:    testSettings(),
		gptProvider: categorizer,
	}
	if err := syncUseCase.categorizeTransactions(t.Context(), testSyncProfileSettings("sync-profile"), transactions); err != nil {
		t.Fatalf("categorizeTransactions() error = %v", err)
	}

	for _, transaction := range transactions {
		if transaction.Category != entity.DefaultFallbackCategory {
			t.Fatalf("transaction = %#v", transaction)
		}
	}
}

func TestSyncUsesSyncProfileSpecificCategorizationSettings(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	settings := entity.SyncSettings{SyncProfiles: []entity.SyncProfileSettings{
		{
			ID:               "first-sync-profile",
			Categories:       []entity.Category{"Alpha", "Others"},
			ColorsByCategory: map[entity.Category]entity.Color{"Alpha": entity.Red, "Others": entity.Gray},
			Mappings:         map[string]entity.Category{"First example": "Alpha"},
			Fallback:         "Others",
		},
		{
			ID:               "second-sync-profile",
			Categories:       []entity.Category{"Beta", "Outros"},
			ColorsByCategory: map[entity.Category]entity.Color{"Beta": entity.Blue, "Outros": entity.Purple},
			Mappings:         map[string]entity.Category{"Second example": "Beta"},
			Fallback:         "Outros",
		},
	}}

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsBySyncProfileID(mock.Anything, mock.Anything, date, date).
		RunAndReturn(func(_ context.Context, syncProfileID string, _, _ time.Time) ([]entity.Transaction, error) {
			return []entity.Transaction{{Name: syncProfileID + " transaction", Amount: 10, Date: date}}, nil
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

			if name == "first-sync-profile transaction" {
				return `{"first-sync-profile transaction":"Beta"}`, nil
			}

			return `{"second-sync-profile transaction":"Alpha"}`, nil
		}).
		Twice()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, mock.Anything).Return(nil, nil).Twice()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, mock.Anything, "Jan 2026").
		RunAndReturn(func(_ context.Context, syncProfileID, title string) (entity.Table, error) {
			return entity.Table{ID: syncProfileID + "-table", Title: title}, nil
		}).
		Twice()

	insertedCategories := make(map[string]entity.Category, 2)
	var insertedMutex sync.Mutex
	store.EXPECT().
		InsertTransaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, syncProfileID, _ string, transaction entity.Transaction) {
			insertedMutex.Lock()
			insertedCategories[syncProfileID] = transaction.Category
			insertedMutex.Unlock()
		}).
		Return(nil).
		Twice()

	err := NewSync(
		testMaxConcurrentOperations,
		settings,
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(context.Background(), SyncInput{StartDate: date, EndDate: date})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	firstInput := inputs["first-sync-profile transaction"]
	if len(firstInput.Categories) != 2 || firstInput.Categories[0] != "Alpha" ||
		firstInput.Mappings["First example"] != "Alpha" || firstInput.Fallback != "Others" {
		t.Fatalf("first categorization input = %#v", firstInput)
	}
	secondInput := inputs["second-sync-profile transaction"]
	if len(secondInput.Categories) != 2 || secondInput.Categories[0] != "Beta" ||
		secondInput.Mappings["Second example"] != "Beta" || secondInput.Fallback != "Outros" {
		t.Fatalf("second categorization input = %#v", secondInput)
	}
	if insertedCategories["first-sync-profile"] != "Others" || insertedCategories["second-sync-profile"] != "Outros" {
		t.Fatalf("inserted categories = %#v", insertedCategories)
	}
}

func TestSyncProcessesEveryMonthAndDeduplicates(t *testing.T) {
	t.Parallel()

	janDate := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)
	febDate := time.Date(2026, time.February, 10, 12, 0, 0, 0, time.UTC)
	existing := entity.Transaction{Name: "Market", Amount: 10, Date: janDate}
	newJanuary := entity.Transaction{Name: "Cafe", Amount: 20, Date: janDate}
	newFebruary := entity.Transaction{Name: "Unknown", Amount: 30, Date: febDate}

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", janDate, febDate).
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
		ListTables(mock.Anything, "sync-profile").
		Return([]entity.Table{{ID: "jan", Title: "Jan 2026"}}, nil).
		Once()
	store.EXPECT().
		ListTransactions(mock.Anything, "sync-profile", "jan").
		Return([]entity.Transaction{existing}, nil).
		Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "sync-profile", "Feb 2026").
		Return(entity.Table{ID: "created-Feb 2026", Title: "Feb 2026"}, nil).
		Once()

	var insertedMutex sync.Mutex
	inserted := make([]insertedTransaction, 0, 2)
	store.EXPECT().
		InsertTransaction(mock.Anything, "sync-profile", mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ string, tableID string, transaction entity.Transaction) {
			insertedMutex.Lock()
			inserted = append(inserted, insertedTransaction{tableID: tableID, transaction: transaction})
			insertedMutex.Unlock()
		}).
		Return(nil).
		Twice()

	syncUseCase := NewSync(
		testMaxConcurrentOperations,
		testSettings("sync-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	)
	err := syncUseCase.Execute(context.Background(), SyncInput{
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

func TestSyncEnrichesUniqueCompany(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	company := mockcompanyapi.NewMockCompanyAPI(t)
	company.EXPECT().
		GetCompanyByID(mock.Anything, "12345678000195").
		Return(entity.Company{TradingName: "Company"}, nil).
		Once()

	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
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
	store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "sync-profile", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(
			mock.Anything,
			"sync-profile",
			"jan",
			mock.MatchedBy(func(transaction entity.Transaction) bool {
				return transaction.Name == "Company" && transaction.Category == "Food"
			}),
		).
		Return(nil).
		Twice()

	if err := NewSync(testMaxConcurrentOperations, testSettings("sync-profile"), company, categorizer, store, source).Execute(
		context.Background(),
		SyncInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestSyncCompanyLookupFailureIsNonFatal(t *testing.T) {
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
		ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
		Return([]entity.Transaction{{Name: "12.345.678/0001-95", Amount: 1, Date: date}}, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		Return(`{"12.345.678/0001-95":"Food"}`, nil).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "sync-profile", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(
			mock.Anything,
			"sync-profile",
			"jan",
			mock.MatchedBy(func(transaction entity.Transaction) bool {
				return transaction.Name == "12.345.678/0001-95"
			}),
		).
		Return(nil).
		Once()

	if err := NewSync(testMaxConcurrentOperations, testSettings("sync-profile"), company, categorizer, store, source).Execute(
		context.Background(),
		SyncInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestSyncEmptyRangeSkipsCategorizer(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
		Return(nil, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "sync-profile", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()

	if err := NewSync(
		testMaxConcurrentOperations,
		testSettings("sync-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		SyncInput{StartDate: date, EndDate: date},
	); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestSyncPropagatesSourceErrorAndCancellation(t *testing.T) {
	t.Parallel()

	date := time.Now()
	t.Run("source error", func(t *testing.T) {
		wantErr := errors.New("source failed")
		source := mockopenfinance.NewMockOpenFinance(t)
		source.EXPECT().
			ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
			Return(nil, wantErr).
			Once()

		err := NewSync(
			testMaxConcurrentOperations,
			testSettings("sync-profile"),
			noCompanyLookup(t),
			mockgpt.NewMockGPT(t),
			mocksheet.NewMockSheet(t),
			source,
		).Execute(context.Background(), SyncInput{StartDate: date, EndDate: date})
		if !errors.Is(err, wantErr) {
			t.Fatalf("Execute() error = %v, want source error", err)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		canceledContext, cancel := context.WithCancel(context.Background())
		cancel()

		source := mockopenfinance.NewMockOpenFinance(t)
		source.EXPECT().
			ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
			RunAndReturn(func(ctx context.Context, _ string, _, _ time.Time) ([]entity.Transaction, error) {
				return nil, ctx.Err()
			}).
			Once()

		err := NewSync(
			testMaxConcurrentOperations,
			testSettings("sync-profile"),
			noCompanyLookup(t),
			mockgpt.NewMockGPT(t),
			mocksheet.NewMockSheet(t),
			source,
		).Execute(canceledContext, SyncInput{StartDate: date, EndDate: date})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Execute() error = %v, want context canceled", err)
		}
	})
}

func TestSyncPropagatesCategorizerAndStoreErrors(t *testing.T) {
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
				ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
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
				store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, listTablesErr).Once()
			case "insert":
				categorizer.EXPECT().
					CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
					Return(`{"Store":"Food"}`, nil).
					Once()
				store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, nil).Once()
				store.EXPECT().
					CreateTransactionsTable(mock.Anything, "sync-profile", "Jan 2026").
					Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
					Once()
				store.EXPECT().
					InsertTransaction(mock.Anything, "sync-profile", "jan", mock.Anything).
					Return(insertErr).
					Once()
			}

			err := NewSync(
				testMaxConcurrentOperations,
				testSettings("sync-profile"),
				noCompanyLookup(t),
				categorizer,
				store,
				source,
			).Execute(
				context.Background(),
				SyncInput{StartDate: date, EndDate: date},
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Execute() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestSyncBoundsSyncProfileConcurrency(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	syncProfileIDs := make([]string, 12)
	for index := range syncProfileIDs {
		syncProfileIDs[index] = fmt.Sprintf("sync-profile-%d", index)
	}

	var activeSources atomic.Int32
	var maximumSources atomic.Int32
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsBySyncProfileID(mock.Anything, mock.Anything, date, date).
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
		Times(len(syncProfileIDs))

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, mock.Anything).Return(nil, nil).Times(len(syncProfileIDs))
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, mock.Anything, "Jan 2026").
		RunAndReturn(func(_ context.Context, syncProfileID, title string) (entity.Table, error) {
			return entity.Table{ID: syncProfileID + "-jan", Title: title}, nil
		}).
		Times(len(syncProfileIDs))

	if err := NewSync(
		testMaxConcurrentOperations,
		testSettings(syncProfileIDs...),
		noCompanyLookup(t),
		mockgpt.NewMockGPT(t),
		store,
		source,
	).Execute(context.Background(), SyncInput{StartDate: date, EndDate: date}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if maximumSources.Load() > testMaxConcurrentOperations {
		t.Fatalf("source concurrency = %d, limit %d", maximumSources.Load(), testMaxConcurrentOperations)
	}
	if maximumSources.Load() < 2 {
		t.Fatalf("expected concurrent execution, got sources=%d", maximumSources.Load())
	}
}

func TestSyncBoundsInsertConcurrency(t *testing.T) {
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
		ListTransactionsBySyncProfileID(mock.Anything, "sync-profile", date, date).
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
	store.EXPECT().ListTables(mock.Anything, "sync-profile").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "sync-profile", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(mock.Anything, "sync-profile", "jan", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ string, _ entity.Transaction) error {
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

	if err := NewSync(
		testMaxConcurrentOperations,
		testSettings("sync-profile"),
		noCompanyLookup(t),
		categorizer,
		store,
		source,
	).Execute(
		context.Background(),
		SyncInput{StartDate: date, EndDate: date},
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
