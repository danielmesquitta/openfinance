package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

const (
	categorizationSystemMessage = "Categorize each transaction name using only the supplied categories. " +
		"Return one JSON object whose keys are the exact transaction names and whose values are category names. " +
		"Use the fallback when uncertain."
	combinedCategorizationSystemMessage = "Classify each transaction name in two dimensions using only the supplied values. " +
		"Category describes what the transaction represents. Budget group describes its financial or budgeting purpose. " +
		"Budget group examples map category names to budget groups; use them as guidance and infer a group for unmapped categories. " +
		"Return one JSON object whose keys are the exact transaction names and whose values are objects with category and budget_group fields. " +
		"Use each dimension's fallback when uncertain."
)

type IngestInput struct {
	StartDate time.Time `validate:"required"`
	EndDate   time.Time `validate:"required,gtefield=StartDate"`
}

type IngestExecutor interface {
	Execute(ctx context.Context, input IngestInput) error
}

type Ingest struct {
	val                     *validator.Validator
	maxConcurrentOperations int
	settings                entity.IngestSettings
	companyAPIProvider      companyapi.APIProvider
	gptProvider             gpt.Provider
	sheetProvider           sheet.Provider
	openFinanceAPIProvider  openfinance.APIProvider
}

func NewIngest(
	val *validator.Validator,
	maxConcurrentOperations int,
	settings entity.IngestSettings,
	companyAPIProvider companyapi.APIProvider,
	gptProvider gpt.Provider,
	sheetProvider sheet.Provider,
	openFinanceAPIProvider openfinance.APIProvider,
) *Ingest {
	return &Ingest{
		val:                     val,
		maxConcurrentOperations: maxConcurrentOperations,
		settings:                settings,
		companyAPIProvider:      companyAPIProvider,
		gptProvider:             gptProvider,
		sheetProvider:           sheetProvider,
		openFinanceAPIProvider:  openFinanceAPIProvider,
	}
}

func (s *Ingest) Execute(ctx context.Context, input IngestInput) error {
	if err := s.val.Validate(input); err != nil {
		return fmt.Errorf("invalid ingest input: %w", err)
	}

	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(s.maxConcurrentOperations)

	for _, ingestProfileSettings := range s.settings.IngestProfiles {
		group.Go(func() error {
			if err := s.ingestProfile(groupContext, ingestProfileSettings, input); err != nil {
				return fmt.Errorf("ingest profile %q: %w", ingestProfileSettings.ID, err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("ingest profiles: %w", err)
	}

	return nil
}

func (s *Ingest) ingestProfile(
	ctx context.Context,
	settings entity.IngestProfileSettings,
	input IngestInput,
) error {
	transactions, err := s.openFinanceAPIProvider.ListTransactionsByIngestProfileID(
		ctx,
		settings.ID,
		input.StartDate,
		input.EndDate,
	)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}
	transactions = filterTransactions(transactions, settings.IgnoreSamePersonTransfers)

	s.enrichTransactionNames(ctx, transactions)

	if err := s.categorizeTransactions(ctx, settings, transactions); err != nil {
		return fmt.Errorf("categorize transactions: %w", err)
	}

	tables, err := s.sheetProvider.ListTables(ctx, settings.ID)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}

	tableByTitle := make(map[string]sheet.Table, len(tables))
	for _, table := range tables {
		tableByTitle[table.Title] = table
	}

	transactionsByMonth := groupTransactionsByMonth(transactions)
	for _, month := range monthsInRange(input.StartDate, input.EndDate) {
		prepared, err := s.prepareTransactionTable(
			ctx,
			settings,
			month,
			tableByTitle,
			transactionsByMonth[newTransactionMonth(month)],
		)
		if err != nil {
			return err
		}

		if err := s.insertTransactions(
			ctx,
			settings.ID,
			prepared.table.ID,
			prepared.language,
			prepared.transactions,
		); err != nil {
			return fmt.Errorf("insert transactions into table %q: %w", prepared.table.Title, err)
		}
	}

	return nil
}

type preparedTransactionTable struct {
	table        sheet.Table
	language     entity.Language
	transactions []entity.Transaction
}

func (s *Ingest) prepareTransactionTable(
	ctx context.Context,
	settings entity.IngestProfileSettings,
	month time.Time,
	tableByTitle map[string]sheet.Table,
	transactions []entity.Transaction,
) (preparedTransactionTable, error) {
	configuredLanguage := normalizedLanguage(settings.Language)
	title := localizedTransactionTableTitle(month, configuredLanguage)
	table, tableLanguage, exists := transactionTableForMonth(tableByTitle, month, configuredLanguage)
	if !exists {
		created, err := s.sheetProvider.CreateTable(
			ctx,
			settings.ID,
			title,
			transactionTableOptions(settings)...,
		)
		if err != nil {
			return preparedTransactionTable{}, fmt.Errorf("create table %q: %w", title, err)
		}

		tableByTitle[title] = created

		return preparedTransactionTable{
			table:        created,
			language:     configuredLanguage,
			transactions: transactions,
		}, nil
	}

	if budgetGroupColumn, enabled := budgetGroupTableColumn(settings, tableLanguage); enabled {
		if err := s.sheetProvider.EnsureTableColumns(
			ctx,
			settings.ID,
			table.ID,
			budgetGroupColumn,
		); err != nil {
			return preparedTransactionTable{}, fmt.Errorf("upgrade table %q: %w", table.Title, err)
		}
	}

	newTransactions, err := s.onlyNewTransactions(
		ctx,
		settings.ID,
		table.ID,
		tableLanguage,
		transactions,
	)
	if err != nil {
		return preparedTransactionTable{}, fmt.Errorf("filter transactions for table %q: %w", table.Title, err)
	}

	return preparedTransactionTable{
		table:        table,
		language:     tableLanguage,
		transactions: newTransactions,
	}, nil
}

func (s *Ingest) enrichTransactionNames(ctx context.Context, transactions []entity.Transaction) {
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

func (s *Ingest) lookupCompanyNames(ctx context.Context, documents []string) map[string]string {
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

func (s *Ingest) lookupCompanyName(ctx context.Context, document string) (string, bool) {
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

func (s *Ingest) categorizeTransactions(
	ctx context.Context,
	settings entity.IngestProfileSettings,
	transactions []entity.Transaction,
) error {
	names := uniqueTransactionNames(transactions)
	if len(names) == 0 {
		return nil
	}
	if len(settings.BudgetGroups) > 0 {
		return s.categorizeTransactionsWithBudgetGroups(ctx, settings, transactions, names)
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

type transactionClassifications struct {
	Category    entity.Category    `json:"category"`
	BudgetGroup entity.BudgetGroup `json:"budget_group"`
}

func (s *Ingest) categorizeTransactionsWithBudgetGroups(
	ctx context.Context,
	settings entity.IngestProfileSettings,
	transactions []entity.Transaction,
	names []string,
) error {
	payload, err := json.Marshal(struct {
		TransactionNames    []string                               `json:"transaction_names"`
		Categories          []entity.Category                      `json:"categories"`
		CategoryMappings    map[string]entity.Category             `json:"category_examples"`
		CategoryFallback    entity.Category                        `json:"category_fallback"`
		BudgetGroups        []entity.BudgetGroup                   `json:"budget_groups"`
		BudgetGroupMappings map[entity.Category]entity.BudgetGroup `json:"budget_group_examples"`
		BudgetGroupFallback entity.BudgetGroup                     `json:"budget_group_fallback"`
	}{
		TransactionNames:    names,
		Categories:          settings.Categories,
		CategoryMappings:    settings.Mappings,
		CategoryFallback:    settings.Fallback,
		BudgetGroups:        settings.BudgetGroups,
		BudgetGroupMappings: settings.BudgetGroupMappings,
		BudgetGroupFallback: settings.BudgetGroupFallback,
	})
	if err != nil {
		return fmt.Errorf("marshal combined categorization input: %w", err)
	}

	content, err := s.gptProvider.CreateChatCompletion(
		ctx,
		string(payload),
		gpt.WithSystemMessage(combinedCategorizationSystemMessage),
		gpt.WithJSONResponse(),
	)
	if err != nil {
		return fmt.Errorf("create combined categorization chat completion: %w", err)
	}

	classificationsByName := make(map[string]transactionClassifications)
	if err := json.Unmarshal([]byte(content), &classificationsByName); err != nil {
		return fmt.Errorf("decode combined categorization chat completion: %w", err)
	}

	allowedCategories := make(map[entity.Category]struct{}, len(settings.Categories))
	for _, category := range settings.Categories {
		allowedCategories[category] = struct{}{}
	}
	allowedBudgetGroups := make(map[entity.BudgetGroup]struct{}, len(settings.BudgetGroups))
	for _, budgetGroup := range settings.BudgetGroups {
		allowedBudgetGroups[budgetGroup] = struct{}{}
	}

	for index := range transactions {
		classifications, ok := classificationsByName[transactions[index].Name]
		if _, allowed := allowedCategories[classifications.Category]; !ok || !allowed {
			classifications.Category = settings.Fallback
		}
		if _, allowed := allowedBudgetGroups[classifications.BudgetGroup]; !ok || !allowed {
			classifications.BudgetGroup = settings.BudgetGroupFallback
		}

		transactions[index].Category = classifications.Category
		transactions[index].BudgetGroup = classifications.BudgetGroup
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

type transactionMonth struct {
	year  int
	month time.Month
}

func newTransactionMonth(value time.Time) transactionMonth {
	return transactionMonth{year: value.Year(), month: value.Month()}
}

func groupTransactionsByMonth(transactions []entity.Transaction) map[transactionMonth][]entity.Transaction {
	transactionsByMonth := make(map[transactionMonth][]entity.Transaction)
	for _, transaction := range transactions {
		month := newTransactionMonth(transaction.Date)
		transactionsByMonth[month] = append(transactionsByMonth[month], transaction)
	}

	return transactionsByMonth
}

func transactionTableForMonth(
	tableByTitle map[string]sheet.Table,
	month time.Time,
	preferredLanguage entity.Language,
) (sheet.Table, entity.Language, bool) {
	preferredLanguage = normalizedLanguage(preferredLanguage)
	preferredTitle := localizedTransactionTableTitle(month, preferredLanguage)
	if table, exists := tableByTitle[preferredTitle]; exists {
		return table, preferredLanguage, true
	}

	alternative := alternateLanguage(preferredLanguage)
	alternativeTitle := localizedTransactionTableTitle(month, alternative)
	if table, exists := tableByTitle[alternativeTitle]; exists {
		return table, alternative, true
	}

	return sheet.Table{}, "", false
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

func (s *Ingest) onlyNewTransactions(
	ctx context.Context,
	ingestProfileID, tableID string,
	language entity.Language,
	transactions []entity.Transaction,
) ([]entity.Transaction, error) {
	rows, err := s.sheetProvider.ListRows(ctx, ingestProfileID, tableID)
	if err != nil {
		return nil, fmt.Errorf("list existing transactions: %w", err)
	}

	existingTransactions := make([]entity.Transaction, 0, len(rows))
	for _, row := range rows {
		transaction, err := rowToTransaction(row, language)
		if err != nil {
			return nil, fmt.Errorf("map existing transaction row: %w", err)
		}
		existingTransactions = append(existingTransactions, transaction)
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

func (s *Ingest) insertTransactions(
	ctx context.Context,
	ingestProfileID, tableID string,
	language entity.Language,
	transactions []entity.Transaction,
) error {
	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(s.maxConcurrentOperations)

	for _, transaction := range transactions {
		group.Go(func() error {
			if err := s.sheetProvider.InsertRow(
				groupContext,
				ingestProfileID,
				tableID,
				transactionToRow(transaction, language),
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

var _ IngestExecutor = (*Ingest)(nil)
