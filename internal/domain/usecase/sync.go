package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

const (
	categorizationSystemMessage = "Categorize each transaction name using only the supplied categories. " +
		"Return one JSON object whose keys are the exact transaction names and whose values are category names. " +
		"Use the fallback when uncertain."
)

type SyncInput struct {
	StartDate time.Time
	EndDate   time.Time
}

type SyncExecutor interface {
	Execute(ctx context.Context, input SyncInput) error
}

func (input SyncInput) Validate() error {
	if input.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if input.EndDate.IsZero() {
		return errors.New("end date is required")
	}

	if input.StartDate.After(input.EndDate) {
		return errors.New("start date cannot be after end date")
	}

	return nil
}

type Sync struct {
	maxConcurrentOperations int
	settings                entity.SyncSettings
	companyAPIProvider      companyapi.APIProvider
	gptProvider             gpt.Provider
	sheetProvider           sheet.Provider
	openFinanceAPIProvider  openfinance.APIProvider
}

func NewSync(
	maxConcurrentOperations int,
	settings entity.SyncSettings,
	companyAPIProvider companyapi.APIProvider,
	gptProvider gpt.Provider,
	sheetProvider sheet.Provider,
	openFinanceAPIProvider openfinance.APIProvider,
) *Sync {
	return &Sync{
		maxConcurrentOperations: maxConcurrentOperations,
		settings:                settings,
		companyAPIProvider:      companyAPIProvider,
		gptProvider:             gptProvider,
		sheetProvider:           sheetProvider,
		openFinanceAPIProvider:  openFinanceAPIProvider,
	}
}

func (s *Sync) Execute(ctx context.Context, input SyncInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("invalid sync input: %w", err)
	}

	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(s.maxConcurrentOperations)

	for _, syncProfileSettings := range s.settings.SyncProfiles {
		group.Go(func() error {
			if err := s.syncProfile(groupContext, syncProfileSettings, input); err != nil {
				return fmt.Errorf("sync profile %q: %w", syncProfileSettings.ID, err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("sync profiles: %w", err)
	}

	return nil
}

func (s *Sync) syncProfile(
	ctx context.Context,
	settings entity.SyncProfileSettings,
	input SyncInput,
) error {
	transactions, err := s.openFinanceAPIProvider.ListTransactionsBySyncProfileID(
		ctx,
		settings.ID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	s.enrichTransactionNames(ctx, transactions)

	if err := s.categorizeTransactions(ctx, settings, transactions); err != nil {
		return fmt.Errorf("categorize transactions: %w", err)
	}

	tables, err := s.sheetProvider.ListTables(ctx, settings.ID)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}

	tableByTitle := make(map[string]entity.Table, len(tables))
	for _, table := range tables {
		tableByTitle[table.Title] = table
	}

	transactionsByMonth := groupTransactionsByMonth(transactions)
	for _, month := range monthsInRange(input.StartDate, input.EndDate) {
		title := month.Format("Jan 2006")
		monthTransactions := transactionsByMonth[title]

		table, exists := tableByTitle[title]
		if !exists {
			table, err = s.sheetProvider.CreateTransactionsTable(ctx, settings.ID, title)
			if err != nil {
				return fmt.Errorf("create table %q: %w", title, err)
			}

			tableByTitle[title] = table
		} else {
			monthTransactions, err = s.onlyNewTransactions(ctx, settings.ID, table.ID, monthTransactions)
			if err != nil {
				return fmt.Errorf("filter transactions for table %q: %w", title, err)
			}
		}

		if err := s.insertTransactions(ctx, settings.ID, table.ID, monthTransactions); err != nil {
			return fmt.Errorf("insert transactions into table %q: %w", title, err)
		}
	}

	return nil
}

func (s *Sync) enrichTransactionNames(ctx context.Context, transactions []entity.Transaction) {
	documents := uniqueCompanyDocuments(transactions)
	companyNameByDocument := s.lookupCompanyNames(ctx, documents)
	applyCompanyNames(transactions, companyNameByDocument)
}

func uniqueCompanyDocuments(transactions []entity.Transaction) []string {
	documents := make([]string, 0)
	seenDocuments := make(map[string]struct{})
	for _, transaction := range transactions {
		if !docutil.IsCNPJ(transaction.Name) {
			continue
		}

		document := docutil.CleanDocument(transaction.Name)
		if _, exists := seenDocuments[document]; exists {
			continue
		}

		seenDocuments[document] = struct{}{}
		documents = append(documents, document)
	}

	return documents
}

func (s *Sync) lookupCompanyNames(ctx context.Context, documents []string) map[string]string {
	companyNameByDocument := make(map[string]string, len(documents))
	var mutex sync.Mutex
	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(s.maxConcurrentOperations)

	for _, document := range documents {
		group.Go(func() error {
			name, ok := s.lookupCompanyName(groupContext, document)
			if !ok {
				return nil
			}

			mutex.Lock()
			companyNameByDocument[document] = name
			mutex.Unlock()

			return nil
		})
	}

	_ = group.Wait()

	return companyNameByDocument
}

func (s *Sync) lookupCompanyName(ctx context.Context, document string) (string, bool) {
	company, err := s.companyAPIProvider.GetCompanyByID(ctx, document)
	if err != nil {
		slog.Error("failed to get company", "document", document, "error", err)

		return "", false
	}

	if company.TradingName != "" {
		return company.TradingName, true
	}

	return company.Name, company.Name != ""
}

func applyCompanyNames(transactions []entity.Transaction, companyNameByDocument map[string]string) {
	for index := range transactions {
		document := docutil.CleanDocument(transactions[index].Name)
		if name, ok := companyNameByDocument[document]; ok {
			transactions[index].Name = name
		}
	}
}

func (s *Sync) categorizeTransactions(
	ctx context.Context,
	settings entity.SyncProfileSettings,
	transactions []entity.Transaction,
) error {
	names := uniqueTransactionNames(transactions)
	if len(names) == 0 {
		return nil
	}

	payload, err := json.Marshal(struct {
		TransactionNames []string                   `json:"transaction_names"`
		Categories       []entity.Category          `json:"categories"`
		Mappings         map[string]entity.Category `json:"examples"`
		Fallback         entity.Category            `json:"fallback"`
	}{
		TransactionNames: names,
		Categories:       settings.Categories,
		Mappings:         settings.Mappings,
		Fallback:         settings.Fallback,
	})
	if err != nil {
		return fmt.Errorf("marshal categorization input: %w", err)
	}

	content, err := s.gptProvider.CreateChatCompletion(
		ctx,
		string(payload),
		gpt.WithSystemMessage(categorizationSystemMessage),
		gpt.WithJSONResponse(),
	)
	if err != nil {
		return fmt.Errorf("create categorization chat completion: %w", err)
	}

	categoryByName := make(map[string]entity.Category)
	if err := json.Unmarshal([]byte(content), &categoryByName); err != nil {
		return fmt.Errorf("decode categorization chat completion: %w", err)
	}

	allowedCategories := make(map[entity.Category]struct{}, len(settings.Categories))
	for _, category := range settings.Categories {
		allowedCategories[category] = struct{}{}
	}

	for index := range transactions {
		category, ok := categoryByName[transactions[index].Name]
		if _, allowed := allowedCategories[category]; !ok || !allowed {
			category = settings.Fallback
		}

		transactions[index].Category = category
	}

	return nil
}

func uniqueTransactionNames(transactions []entity.Transaction) []string {
	names := make([]string, 0, len(transactions))
	seen := make(map[string]struct{}, len(transactions))
	for _, transaction := range transactions {
		if _, exists := seen[transaction.Name]; exists {
			continue
		}

		seen[transaction.Name] = struct{}{}
		names = append(names, transaction.Name)
	}

	return names
}

func groupTransactionsByMonth(transactions []entity.Transaction) map[string][]entity.Transaction {
	transactionsByMonth := make(map[string][]entity.Transaction)
	for _, transaction := range transactions {
		month := transaction.Date.Format("Jan 2006")
		transactionsByMonth[month] = append(transactionsByMonth[month], transaction)
	}

	return transactionsByMonth
}

func monthsInRange(startDate, endDate time.Time) []time.Time {
	start := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location())

	months := make([]time.Time, 0, 1)
	for month := start; !month.After(end); month = month.AddDate(0, 1, 0) {
		months = append(months, month)
	}

	return months
}

func (s *Sync) onlyNewTransactions(
	ctx context.Context,
	syncProfileID, tableID string,
	transactions []entity.Transaction,
) ([]entity.Transaction, error) {
	existingTransactions, err := s.sheetProvider.ListTransactions(ctx, syncProfileID, tableID)
	if err != nil {
		return nil, fmt.Errorf("list existing transactions: %w", err)
	}

	seen := make(map[string]struct{}, len(existingTransactions)+len(transactions))
	for _, transaction := range existingTransactions {
		seen[transaction.ID()] = struct{}{}
	}

	newTransactions := make([]entity.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		id := transaction.ID()
		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		newTransactions = append(newTransactions, transaction)
	}

	return newTransactions, nil
}

func (s *Sync) insertTransactions(
	ctx context.Context,
	syncProfileID, tableID string,
	transactions []entity.Transaction,
) error {
	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(s.maxConcurrentOperations)

	for _, transaction := range transactions {
		group.Go(func() error {
			if err := s.sheetProvider.InsertTransaction(
				groupContext,
				syncProfileID,
				tableID,
				transaction,
			); err != nil {
				return fmt.Errorf("insert transaction %q: %w", transaction.ID(), err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("insert transaction batch: %w", err)
	}

	return nil
}

var _ SyncExecutor = (*Sync)(nil)
