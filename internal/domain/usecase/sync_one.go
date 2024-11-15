package usecase

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sourcegraph/conc/iter"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"github.com/danielmesquitta/openfinance/internal/pkg/jsonutil"
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
		return errs.New(err)
	}

	iter.ForEachIdx(transactions, func(i int, t *entity.Transaction) {
		if !docutil.IsCNPJ(t.Name) {
			return
		}
		document := docutil.CleanDocument(t.Name)
		company, err := so.companyAPIProvider.GetCompanyByID(document)
		if err != nil {
			slog.Error("failed to get company by document",
				"document", document,
				"error", err,
			)
			return
		}
		transactions[i].Name = cmp.Or(
			company.TradingName,
			company.Name,
			transactions[i].Name,
		)
	})

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
		return errs.New(err)
	}

	gptMessage := fmt.Sprintf(
		`Read the text below and return in JSON format,
     with key as the transaction name and value as the category name.
     Here is an example response:
     {
       "TAPAJOS EMPREENDIMENTOS IMOBILIARIOS LTDA": "Real state",
       "GROWTH SUPPLEMENTS": "Health and fitness",
       "ALGAR TELECOM": "Telecommunications",
       "Uber *Uber *Trip": "Transportation",
       "99app *99app": "Transportation",
       "ESTADO DE MINAS GERAIS": "Taxes",
       "RECEITA FEDERAL": "Taxes",
       "CEMIG D": "Energy"
     }, return "%s" for unknown categories.
     Be direct and return only the JSON
     %s
    `,
		sheet.CategoryUnknown,
		jsonBytes,
	)

	rawResponse, err := so.gptProvider.CreateChatCompletion(gptMessage)
	if err != nil {
		return errs.New(err)
	}

	jsonResponse := jsonutil.ExtractJSONFromText(rawResponse)
	if len(jsonResponse) != 1 {
		return errs.New("invalid JSON response")
	}

	categoryByTransaction := map[string]string{}
	if err := json.Unmarshal([]byte(jsonResponse[0]), &categoryByTransaction); err != nil {
		return errs.New(err)
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
		return errs.New(err)
	}

	// @TODO: Do a batch insert
	for _, transaction := range transactions {
		_, err := so.sheetProvider.InsertTransaction(
			userID,
			newTableResponse.ID,
			transaction,
		)
		if err != nil {
			slog.Error(
				"failed to insert transaction into database",
				"error",
				err,
			)
		}
	}

	return nil
}
