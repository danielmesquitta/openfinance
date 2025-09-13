package usecase

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"github.com/danielmesquitta/openfinance/internal/pkg/jsonutil"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type SyncOne struct {
	env                    *config.Env
	val                    *validator.Validator
	companyAPIProvider     companyapi.APIProvider
	gptProvider            gpt.Provider
	sheetProvider          sheet.Provider
	openFinanceAPIProvider openfinance.APIProvider
}

func NewSyncOne(
	env *config.Env,
	val *validator.Validator,
	companyAPIProvider companyapi.APIProvider,
	gptProvider gpt.Provider,
	sheetProvider sheet.Provider,
	openFinanceAPIProvider openfinance.APIProvider,
) *SyncOne {
	return &SyncOne{
		env:                    env,
		val:                    val,
		companyAPIProvider:     companyAPIProvider,
		gptProvider:            gptProvider,
		sheetProvider:          sheetProvider,
		openFinanceAPIProvider: openFinanceAPIProvider,
	}
}

func (so *SyncOne) Execute(
	ctx context.Context,
	userID string,
	dto SyncDTO,
) error {
	setDefaultValues(&dto)

	if err := so.val.Validate(dto); err != nil {
		return fmt.Errorf("failed to validate dto: %w", err)
	}

	startDate, endDate, err := parseDates(dto)
	if err != nil {
		return fmt.Errorf("failed to parse dates: %w", err)
	}

	transactions, err := so.fetchTransactions(ctx, userID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	if err := so.enrichTransactionNames(ctx, transactions); err != nil {
		return fmt.Errorf("failed to enrich transaction names: %w", err)
	}

	if err := so.categorizeTransactions(transactions); err != nil {
		return fmt.Errorf("failed to categorize transactions: %w", err)
	}

	if err := so.createAndPopulateTable(ctx, userID, transactions, startDate); err != nil {
		return fmt.Errorf("failed to create and populate table: %w", err)
	}

	return nil
}

func (so *SyncOne) fetchTransactions(
	ctx context.Context,
	userID string,
	startDate, endDate time.Time,
) ([]entity.Transaction, error) {
	transactions, err := so.openFinanceAPIProvider.ListTransactionsByUserID(
		ctx,
		userID,
		startDate,
		endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions by user id: %w", err)
	}

	return transactions, nil
}

func (so *SyncOne) enrichTransactionNames(
	ctx context.Context,
	transactions []entity.Transaction,
) error {
	mu := sync.Mutex{}
	g, _ := errgroup.WithContext(ctx)

	for i, t := range transactions {
		g.Go(func() error {
			return so.enrichSingleTransactionName(t, i, transactions, &mu)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed while getting company by document: %w", err)
	}

	return nil
}

func (so *SyncOne) enrichSingleTransactionName(
	transaction entity.Transaction,
	index int,
	transactions []entity.Transaction,
	mu *sync.Mutex,
) error {
	if !docutil.IsCNPJ(transaction.Name) {
		return nil
	}

	document := docutil.CleanDocument(transaction.Name)
	company, err := so.companyAPIProvider.GetCompanyByID(document)
	if err != nil {
		slog.Error("failed to get company by document",
			"document", document,
			"error", err,
		)
		return nil
	}

	mu.Lock()
	transactions[index].Name = cmp.Or(
		company.TradingName,
		company.Name,
		transactions[index].Name,
	)
	mu.Unlock()

	return nil
}

func (so *SyncOne) categorizeTransactions(transactions []entity.Transaction) error {
	transactionNames := so.extractUniqueTransactionNames(transactions)

	categoryByTransaction, err := so.getCategoriesFromGPT(transactionNames)
	if err != nil {
		return fmt.Errorf("failed to get categories from GPT: %w", err)
	}

	so.applyCategoriesToTransactions(transactions, categoryByTransaction)

	return nil
}

func (*SyncOne) extractUniqueTransactionNames(transactions []entity.Transaction) []string {
	uniqueTransactionNames := map[string]struct{}{}
	transactionNames := []string{}
	for _, t := range transactions {
		if _, ok := uniqueTransactionNames[t.Name]; ok {
			continue
		}
		uniqueTransactionNames[t.Name] = struct{}{}
		transactionNames = append(transactionNames, t.Name)
	}

	return transactionNames
}

func (so *SyncOne) getCategoriesFromGPT(transactionNames []string) (map[string]string, error) {
	jsonBytes, err := json.Marshal(transactionNames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transactions: %w", err)
	}

	gptMessage := fmt.Sprintf(
		`Read the text below and return in JSON format,
     with key as the transaction name and value as the category name.
     Use the categories from the following list: %s
     Here is an example response:
     %s
     Return "%s" for unknown categories.
     Be direct and return only the JSON.
     %s
    `,
		so.env.JSONCategories,
		so.env.JSONMappings,
		entity.CategoryUnknown,
		jsonBytes,
	)

	rawResponse, err := so.gptProvider.CreateChatCompletion(gptMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	const expectedJSONResponseCount = 1
	jsonResponse := jsonutil.ExtractJSONFromText(rawResponse)
	if len(jsonResponse) != expectedJSONResponseCount {
		return nil, errors.New("invalid JSON response")
	}

	categoryByTransaction := map[string]string{}
	if err := json.Unmarshal([]byte(jsonResponse[0]), &categoryByTransaction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return categoryByTransaction, nil
}

func (*SyncOne) applyCategoriesToTransactions(
	transactions []entity.Transaction,
	categoryByTransaction map[string]string,
) {
	for i, t := range transactions {
		category, ok := categoryByTransaction[t.Name]
		if !ok {
			category = string(entity.CategoryUnknown)
		}
		transactions[i].Category = entity.Category(category)
	}
}

func (so *SyncOne) createAndPopulateTable(
	ctx context.Context,
	userID string,
	transactions []entity.Transaction,
	startDate time.Time,
) error {
	title := startDate.Format("Jan 2006")

	newTableResponse, err := so.sheetProvider.CreateTransactionsTable(
		ctx,
		userID,
		title,
	)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	if err := so.insertTransactions(ctx, userID, newTableResponse.ID, transactions); err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}

	return nil
}

func (so *SyncOne) insertTransactions(
	ctx context.Context,
	userID, tableID string,
	transactions []entity.Transaction,
) error {
	// TODO: Do a batch insert
	g, gCtx := errgroup.WithContext(ctx)
	for _, transaction := range transactions {
		g.Go(func() error {
			_, err := so.sheetProvider.InsertTransaction(
				gCtx,
				userID,
				tableID,
				transaction,
			)
			if err != nil {
				return fmt.Errorf("failed to insert transaction: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed while inserting transactions: %w", err)
	}

	return nil
}
