package usecase

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type SyncOne struct {
	val                    *validator.Validator
	companyAPIProvider     companyapi.APIProvider
	gptProvider            gpt.Provider
	sheetProvider          sheet.Provider
	openFinanceAPIProvider openfinance.APIProvider
}

func NewSyncOne(
	val *validator.Validator,
	companyAPIProvider companyapi.APIProvider,
	gptProvider gpt.Provider,
	sheetProvider sheet.Provider,
	openFinanceAPIProvider openfinance.APIProvider,
) *SyncOne {
	return &SyncOne{
		val:                    val,
		companyAPIProvider:     companyAPIProvider,
		gptProvider:            gptProvider,
		sheetProvider:          sheetProvider,
		openFinanceAPIProvider: openFinanceAPIProvider,
	}
}

func (so *SyncOne) Execute(
	userID string,
	dto SyncDTO,
) error {
	setDefaultValues(&dto)

	if err := so.val.Validate(dto); err != nil {
		return err
	}

	startDate, endDate, err := parseDates(dto)
	if err != nil {
		return err
	}

	transactions, err := so.openFinanceAPIProvider.ListTransactionsByUserID(
		userID,
		startDate,
		endDate,
	)

	if err != nil {
		return entity.NewErr(err)
	}

	jobsCount := len(transactions)
	errCh := make(chan error, jobsCount)
	wg := sync.WaitGroup{}
	wg.Add(jobsCount)

	for i, t := range transactions {
		go func() {
			defer wg.Done()
			if !docutil.IsCNPJ(t.Name) {
				return
			}

			document := docutil.CleanDocument(t.Name)

			company, err := so.companyAPIProvider.GetCompanyByID(document)
			if err != nil {
				errCh <- err
				return
			}

			transactions[i].Name = cmp.Or(
				company.TradingName,
				company.Name,
				transactions[i].Name,
			)
		}()
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		err = errors.Join(err, e)
	}

	if err != nil {
		return entity.NewErr(err)
	}

	uniqueTransactionNames := map[string]struct{}{}
	for _, t := range transactions {
		uniqueTransactionNames[t.Name] = struct{}{}
	}

	transactionNames := []string{}
	for name := range uniqueTransactionNames {
		if name == "" {
			continue
		}
		transactionNames = append(transactionNames, name)
	}

	jsonBytes, err := json.Marshal(transactionNames)
	if err != nil {
		return entity.NewErr(err)
	}

	gptMessage := fmt.Sprintf(
		`Give me a JSON hash map, with key being a transaction and value being a category.
     The values should be unique categories for the following transactions: %s
     Here is an example response {
       "TAPAJOS EMPREENDIMENTOS IMOBILIARIOS LTDA": "Real state",
       "GROWTH SUPPLEMENTS": "Health and fitness",
       "ALGAR TELECOM": "Telecommunications",
       "Uber *Uber *Trip": "Transportation",
       "ESTADO DE MINAS GERAIS": "Taxes",
       "CEMIG D": "Energy"
     }, return "%s" for unknown categories.
     Be direct and return only the JSON
    `,
		jsonBytes,
		sheet.CategoryUnknown,
	)

	rawResponse, err := so.gptProvider.CreateChatCompletion(gptMessage)
	if err != nil {
		return entity.NewErr(err)
	}

	jsonResponse, err := fmtRawGPTResponseToJSON(rawResponse)
	if err != nil {
		return entity.NewErr(err)
	}

	categoryByTransaction := map[string]string{}
	if err := json.Unmarshal([]byte(jsonResponse), &categoryByTransaction); err != nil {
		return entity.NewErr(err)
	}

	uniqueCategories := map[string]struct{}{}
	for i, t := range transactions {
		category, ok := categoryByTransaction[t.Name]
		if !ok {
			category = string(sheet.CategoryUnknown)
		}

		transactions[i].Category = category
		uniqueCategories[category] = struct{}{}
	}

	categories := []sheet.Category{}
	for category := range uniqueCategories {
		categories = append(categories, sheet.Category(category))
	}

	year, month, _ := startDate.Date()

	monthAbbreviation := month.String()[0:3]
	title := fmt.Sprintf("%s %d", monthAbbreviation, year)

	newTableResponse, err := so.sheetProvider.CreateTransactionsTable(
		userID,
		sheet.CreateTransactionsTableDTO{
			Title:      title,
			Categories: categories,
		},
	)
	if err != nil {
		return entity.NewErr(err)
	}

	jobsCount = len(transactions)
	errCh = make(chan error)

	for _, transaction := range transactions {
		go func() {
			_, err := so.sheetProvider.InsertTransaction(
				userID,
				newTableResponse.ID,
				transaction,
			)
			errCh <- err
		}()
	}

	for i := 0; i < jobsCount; i++ {
		err = errors.Join(err, <-errCh)
	}

	close(errCh)

	if err != nil {
		return entity.NewErr(err)
	}

	return nil
}

func fmtRawGPTResponseToJSON(s string) (string, error) {
	split := strings.Split(s, "{")
	if len(split) != 2 {
		return "", entity.NewErr("invalid GPT response")
	}

	split = strings.Split(split[1], "}")
	if len(split) != 2 {
		return "", entity.NewErr("invalid GPT response")
	}

	return fmt.Sprintf("{%s}", split[0]), nil
}
