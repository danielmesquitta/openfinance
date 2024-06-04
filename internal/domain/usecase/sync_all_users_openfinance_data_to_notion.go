package usecase

import (
	"fmt"
	"sync"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/dateutil"
	"github.com/danielmesquitta/openfinance/pkg/validator"
)

type SyncAllUsersOpenFinanceDataToNotionUseCase struct {
	val         *validator.Validator
	crypto      crypto.Encrypter
	settingRepo repo.SettingRepo
}

func NewSyncAllUsersOpenFinanceDataToNotionUseCase(
	val *validator.Validator,
	crypto crypto.Encrypter,
	settingRepo repo.SettingRepo,
) *SyncAllUsersOpenFinanceDataToNotionUseCase {
	return &SyncAllUsersOpenFinanceDataToNotionUseCase{
		val:         val,
		crypto:      crypto,
		settingRepo: settingRepo,
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

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := []error{}

	wg.Add(len(userSettings))
	for _, setting := range userSettings {
		go func() {
			defer wg.Done()

			if err := uc.syncUserOpenFinanceDataToNotion(
				setting,
				startDate,
				endDate,
			); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf(
			"error syncing open finance data to notion:\n%+v",
			errs,
		)
	}

	return nil
}

func (uc *SyncAllUsersOpenFinanceDataToNotionUseCase) setDefaultValues(
	dto *SyncAllUsersOpenFinanceDataToNotionDTO,
) {
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
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

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := []error{}

	transactions := []entity.Transaction{}

	wg.Add(len(setting.MeuPluggyAccountIDs))
	for _, accountID := range setting.MeuPluggyAccountIDs {
		go func() {
			defer wg.Done()

			accountTransactions, err := pluggyClient.ListTransactions(
				accountID,
				startDate,
				endDate,
			)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			transactions = append(
				transactions,
				accountTransactions...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error fetching transactions:\n%+v", errs)
	}

	uniqueCategories := map[string]struct{}{}
	for _, transaction := range transactions {
		uniqueCategories[transaction.Category] = struct{}{}
	}

	categories := []string{}
	for category := range uniqueCategories {
		if category == "" {
			continue
		}
		categories = append(categories, category)
	}

	year, month, _ := startDate.Date()
	strMonth := dateutil.MonthMapper[month]
	title := fmt.Sprintf("%s %d", strMonth, year)

	newTableResponse, err := notionClient.NewTable(sheet.NewTableDTO{
		ParentID:   setting.NotionPageID,
		Title:      title,
		Categories: categories,
	})
	if err != nil {
		return fmt.Errorf("error creating new table: %w", err)
	}

	wg.Add(len(transactions))
	for _, transaction := range transactions {
		go func() {
			defer wg.Done()
			if _, err := notionClient.
				InsertRow(newTableResponse.ID, transaction); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error fetching transactions:\n%+v", errs)
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
