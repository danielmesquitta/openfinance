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

func testSettings(userIDs ...string) entity.SyncSettings {
	return entity.SyncSettings{
		UserIDs:    userIDs,
		Categories: []entity.Category{"Food", entity.CategoryUnknown},
		Mappings:   map[string]entity.Category{"Market": "Food"},
		Fallback:   entity.CategoryUnknown,
	}
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
	if err := syncUseCase.categorizeTransactions(t.Context(), transactions); err != nil {
		t.Fatalf("categorizeTransactions() error = %v", err)
	}

	for _, transaction := range transactions {
		if transaction.Category != entity.CategoryUnknown {
			t.Fatalf("transaction = %#v", transaction)
		}
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
		ListTransactionsByUserID(mock.Anything, "user", janDate, febDate).
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
			if input.Mappings["Market"] != "Food" || input.Fallback != entity.CategoryUnknown {
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
		ListTables(mock.Anything, "user").
		Return([]entity.Table{{ID: "jan", Title: "Jan 2026"}}, nil).
		Once()
	store.EXPECT().
		ListTransactions(mock.Anything, "user", "jan").
		Return([]entity.Transaction{existing}, nil).
		Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "user", "Feb 2026").
		Return(entity.Table{ID: "created-Feb 2026", Title: "Feb 2026"}, nil).
		Once()

	var insertedMutex sync.Mutex
	inserted := make([]insertedTransaction, 0, 2)
	store.EXPECT().
		InsertTransaction(mock.Anything, "user", mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ string, tableID string, transaction entity.Transaction) {
			insertedMutex.Lock()
			inserted = append(inserted, insertedTransaction{tableID: tableID, transaction: transaction})
			insertedMutex.Unlock()
		}).
		Return(nil).
		Twice()

	syncUseCase := NewSync(testMaxConcurrentOperations, testSettings("user"), noCompanyLookup(t), categorizer, store, source)
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
		inserted[1].transaction.Category != entity.CategoryUnknown {
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
		ListTransactionsByUserID(mock.Anything, "user", date, date).
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
	store.EXPECT().ListTables(mock.Anything, "user").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "user", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(
			mock.Anything,
			"user",
			"jan",
			mock.MatchedBy(func(transaction entity.Transaction) bool {
				return transaction.Name == "Company" && transaction.Category == "Food"
			}),
		).
		Return(nil).
		Twice()

	if err := NewSync(testMaxConcurrentOperations, testSettings("user"), company, categorizer, store, source).Execute(
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
		ListTransactionsByUserID(mock.Anything, "user", date, date).
		Return([]entity.Transaction{{Name: "12.345.678/0001-95", Amount: 1, Date: date}}, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	categorizer.EXPECT().
		CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
		Return(`{"12.345.678/0001-95":"Food"}`, nil).
		Once()

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "user").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "user", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(
			mock.Anything,
			"user",
			"jan",
			mock.MatchedBy(func(transaction entity.Transaction) bool {
				return transaction.Name == "12.345.678/0001-95"
			}),
		).
		Return(nil).
		Once()

	if err := NewSync(testMaxConcurrentOperations, testSettings("user"), company, categorizer, store, source).Execute(
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
		ListTransactionsByUserID(mock.Anything, "user", date, date).
		Return(nil, nil).
		Once()

	categorizer := mockgpt.NewMockGPT(t)
	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, "user").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "user", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()

	if err := NewSync(testMaxConcurrentOperations, testSettings("user"), noCompanyLookup(t), categorizer, store, source).Execute(
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
			ListTransactionsByUserID(mock.Anything, "user", date, date).
			Return(nil, wantErr).
			Once()

		err := NewSync(
			testMaxConcurrentOperations,
			testSettings("user"),
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
			ListTransactionsByUserID(mock.Anything, "user", date, date).
			RunAndReturn(func(ctx context.Context, _ string, _, _ time.Time) ([]entity.Transaction, error) {
				return nil, ctx.Err()
			}).
			Once()

		err := NewSync(
			testMaxConcurrentOperations,
			testSettings("user"),
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
				ListTransactionsByUserID(mock.Anything, "user", date, date).
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
				store.EXPECT().ListTables(mock.Anything, "user").Return(nil, listTablesErr).Once()
			case "insert":
				categorizer.EXPECT().
					CreateChatCompletion(mock.Anything, mock.Anything, mock.Anything).
					Return(`{"Store":"Food"}`, nil).
					Once()
				store.EXPECT().ListTables(mock.Anything, "user").Return(nil, nil).Once()
				store.EXPECT().
					CreateTransactionsTable(mock.Anything, "user", "Jan 2026").
					Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
					Once()
				store.EXPECT().
					InsertTransaction(mock.Anything, "user", "jan", mock.Anything).
					Return(insertErr).
					Once()
			}

			err := NewSync(testMaxConcurrentOperations, testSettings("user"), noCompanyLookup(t), categorizer, store, source).Execute(
				context.Background(),
				SyncInput{StartDate: date, EndDate: date},
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Execute() error = %v, want %q", err, test.wantErr)
			}
		})
	}
}

func TestSyncBoundsUserConcurrency(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	userIDs := make([]string, 12)
	for index := range userIDs {
		userIDs[index] = fmt.Sprintf("user-%d", index)
	}

	var activeSources atomic.Int32
	var maximumSources atomic.Int32
	source := mockopenfinance.NewMockOpenFinance(t)
	source.EXPECT().
		ListTransactionsByUserID(mock.Anything, mock.Anything, date, date).
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
		Times(len(userIDs))

	store := mocksheet.NewMockSheet(t)
	store.EXPECT().ListTables(mock.Anything, mock.Anything).Return(nil, nil).Times(len(userIDs))
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, mock.Anything, "Jan 2026").
		RunAndReturn(func(_ context.Context, userID, title string) (entity.Table, error) {
			return entity.Table{ID: userID + "-jan", Title: title}, nil
		}).
		Times(len(userIDs))

	if err := NewSync(
		testMaxConcurrentOperations,
		testSettings(userIDs...),
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
		ListTransactionsByUserID(mock.Anything, "user", date, date).
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
	store.EXPECT().ListTables(mock.Anything, "user").Return(nil, nil).Once()
	store.EXPECT().
		CreateTransactionsTable(mock.Anything, "user", "Jan 2026").
		Return(entity.Table{ID: "jan", Title: "Jan 2026"}, nil).
		Once()
	store.EXPECT().
		InsertTransaction(mock.Anything, "user", "jan", mock.Anything).
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

	if err := NewSync(testMaxConcurrentOperations, testSettings("user"), noCompanyLookup(t), categorizer, store, source).Execute(
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
