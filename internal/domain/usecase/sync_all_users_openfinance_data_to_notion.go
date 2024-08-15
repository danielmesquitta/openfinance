package usecase

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/paemuri/brdoc"
)

type SyncAllUsersOpenFinanceDataToNotionUseCase struct {
	val         *validator.Validator
	crypto      crypto.Encrypter
	settingRepo repo.SettingRepo
	companyAPI  companyapi.API
	gptProvider gpt.GPTProvider
}

func NewSyncAllUsersOpenFinanceDataToNotionUseCase(
	val *validator.Validator,
	crypto crypto.Encrypter,
	settingRepo repo.SettingRepo,
	companyAPI companyapi.API,
	gptProvider gpt.GPTProvider,
) *SyncAllUsersOpenFinanceDataToNotionUseCase {
	return &SyncAllUsersOpenFinanceDataToNotionUseCase{
		val:         val,
		crypto:      crypto,
		settingRepo: settingRepo,
		companyAPI:  companyAPI,
		gptProvider: gptProvider,
	}
}

type SyncAllUsersOpenFinanceDataToNotionDTO struct {
	StartDate string `validate:"datetime=2006-01-02T15:04:05Z07:00" json:"start_date,omitempty"`
	EndDate   string `validate:"datetime=2006-01-02T15:04:05Z07:00" json:"end_date,omitempty"`
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) Execute(
	dto SyncAllUsersOpenFinanceDataToNotionDTO,
) error {
	uc.setDefaultValues(&dto)

	if err := uc.val.Validate(dto); err != nil {
		return err
	}

	startDate, endDate, err := uc.parseDates(dto)
	if err != nil {
		return err
	}

	userSettings, err := uc.settingRepo.ListSettings()
	if err != nil {
		return fmt.Errorf("error listing settings: %w", err)
	}

	jobsCount := len(userSettings)
	errCh := make(chan error)

	for _, setting := range userSettings {
		go func() {
			err := uc.syncUserOpenFinanceDataToNotion(
				setting,
				startDate,
				endDate,
			)
			errCh <- err
		}()
	}

	for i := 0; i < jobsCount; i++ {
		err = errors.Join(err, <-errCh)
	}

	close(errCh)

	if err != nil {
		return fmt.Errorf(
			"error syncing open finance data to notion:\n%w",
			err,
		)
	}

	return nil
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) setDefaultValues(
	dto *SyncAllUsersOpenFinanceDataToNotionDTO,
) {
	now := time.Now()
	startOfMonth := time.Date(
		now.Year(),
		now.Month()-1,
		1,
		0,
		0,
		0,
		0,
		time.Local,
	) // day 1 of previous month
	endOfMonth := startOfMonth.AddDate(0, 1, -1) // last day of previous month
	if dto.StartDate == "" {
		dto.StartDate = startOfMonth.Format(time.RFC3339)
	}
	if dto.EndDate == "" {
		dto.EndDate = endOfMonth.Format(time.RFC3339)
	}
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) parseDates(
	dto SyncAllUsersOpenFinanceDataToNotionDTO,
) (startDate time.Time, endDate time.Time, err error) {
	startDate, err = time.Parse(time.RFC3339, dto.StartDate)
	if err != nil {
		appErr := entity.ErrValidation
		appErr.Message = "invalid start date"
		return time.Time{}, time.Time{}, appErr
	}

	endDate, err = time.Parse(time.RFC3339, dto.EndDate)
	if err != nil {
		appErr := entity.ErrValidation
		appErr.Message = "invalid end date"
		return time.Time{}, time.Time{}, appErr
	}

	return startDate, endDate, nil
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) syncUserOpenFinanceDataToNotion(
	setting entity.Setting,
	startDate, endDate time.Time,
) error {
	if err := uc.decryptSetting(&setting); err != nil {
		return fmt.Errorf("error decrypting setting values: %w", err)
	}

	pluggyClient := meupluggyapi.NewClient(
		setting.MeuPluggyClientID,
		setting.MeuPluggyClientSecret,
	)

	notionClient := notionapi.NewClient(setting.NotionToken)

	jobsCount := len(setting.MeuPluggyAccountIDs)
	wg := &sync.WaitGroup{}
	wg.Add(jobsCount)
	transactionsCh := make(chan []entity.Transaction, jobsCount)
	errCh := make(chan error, jobsCount)

	for _, accountID := range setting.MeuPluggyAccountIDs {
		go func() {
			defer wg.Done()

			transactions, err := pluggyClient.ListTransactions(
				accountID,
				startDate,
				endDate,
			)
			if err != nil {
				errCh <- err
				return
			}

			transactionsCh <- transactions
		}()
	}

	wg.Wait()
	close(errCh)
	close(transactionsCh)

	var err error
	for e := range errCh {
		err = errors.Join(err, e)
	}

	if err != nil {
		return fmt.Errorf("error fetching transactions: %w", err)
	}

	transactions := []entity.Transaction{}
	for t := range transactionsCh {
		transactions = append(transactions, t...)
	}

	jobsCount = len(transactions)
	errCh = make(chan error, jobsCount)
	wg.Add(jobsCount)

	for i, t := range transactions {
		go func() {
			defer wg.Done()
			if !brdoc.IsCNPJ(t.Name) {
				return
			}

			re := regexp.MustCompile("[^0-9]")
			document := re.ReplaceAllString(t.Name, "")

			company, err := uc.companyAPI.GetCompanyByID(document)
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
		return fmt.Errorf("failed to get company data: %w", err)
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
		return fmt.Errorf("error marshalling transaction names: %w", err)
	}

	unknownCategory := "Others"
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
		unknownCategory,
	)

	gptResponse, err := uc.gptProvider.CreateChatCompletion(gptMessage)
	if err != nil {
		return fmt.Errorf("error creating chat completion: %w", err)
	}

	expectedJSON, err := formatGPTResponseToExpectedJSON(gptResponse)
	if err != nil {
		return fmt.Errorf("error formatting GPT response: %w", err)
	}

	categoryByTransaction := map[string]string{}
	if err := json.Unmarshal([]byte(expectedJSON), &categoryByTransaction); err != nil {
		return fmt.Errorf("error unmarshalling GPT response: %w", err)
	}

	uniqueCategories := map[string]struct{}{}
	for i, t := range transactions {
		category, ok := categoryByTransaction[t.Name]
		if !ok {
			category = unknownCategory
		}

		transactions[i].Category = category
		uniqueCategories[category] = struct{}{}
	}

	categories := []string{}
	for category := range uniqueCategories {
		categories = append(categories, category)
	}

	year, month, _ := startDate.Date()

	monthAbbreviation := month.String()[0:3]
	title := fmt.Sprintf("%s %d", monthAbbreviation, year)

	newTableResponse, err := notionClient.NewTable(sheet.NewTableDTO{
		ParentID:   setting.NotionPageID,
		Title:      title,
		Categories: categories,
	})
	if err != nil {
		return fmt.Errorf("error creating new table: %w", err)
	}

	jobsCount = len(transactions)
	errCh = make(chan error)

	for _, transaction := range transactions {
		go func() {
			_, err := notionClient.InsertRow(newTableResponse.ID, transaction)
			errCh <- err
		}()
	}

	for i := 0; i < jobsCount; i++ {
		err = errors.Join(err, <-errCh)
	}

	close(errCh)

	if err != nil {
		return fmt.Errorf("error fetching transactions:\n%w", err)
	}

	return nil
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) decryptSetting(
	setting *entity.Setting,
) error {
	for i, accountID := range setting.MeuPluggyAccountIDs {
		text, err := uc.crypto.Decrypt(accountID)
		if err != nil {
			return err
		}
		setting.MeuPluggyAccountIDs[i] = text
	}

	textMeuPluggyClientID, err := uc.crypto.Decrypt(
		setting.MeuPluggyClientID,
	)
	if err != nil {
		return err
	}
	setting.MeuPluggyClientID = textMeuPluggyClientID

	textMeuPluggyClientSecret, err := uc.crypto.Decrypt(
		setting.MeuPluggyClientSecret,
	)
	if err != nil {
		return err
	}
	setting.MeuPluggyClientSecret = textMeuPluggyClientSecret

	textNotionPageID, err := uc.crypto.Decrypt(setting.NotionPageID)
	if err != nil {
		return err
	}
	setting.NotionPageID = textNotionPageID

	textNotionToken, err := uc.crypto.Decrypt(setting.NotionToken)
	if err != nil {
		return err
	}
	setting.NotionToken = textNotionToken

	return nil
}

func formatGPTResponseToExpectedJSON(s string) (string, error) {
	split := strings.Split(s, "{")
	if len(split) != 2 {
		return "", errors.New("invalid GPT response")
	}

	split = strings.Split(split[1], "}")
	if len(split) != 2 {
		return "", errors.New("invalid GPT response")
	}

	return fmt.Sprintf("{%s}", split[0]), nil
}
