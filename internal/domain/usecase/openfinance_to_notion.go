package usecase

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/dateutil"
	"github.com/danielmesquitta/openfinance/pkg/formatter"
	"github.com/danielmesquitta/openfinance/pkg/validator"
)

type OpenFinanceToNotionUseCase struct {
	e   *config.Env
	v   *validator.Validator
	nac *notionapi.Client
	mac *meupluggyapi.Client
}

func NewOpenFinanceToNotionUseCase(
	e *config.Env,
	v *validator.Validator,
	nac *notionapi.Client,
	mac *meupluggyapi.Client,
) *OpenFinanceToNotionUseCase {
	return &OpenFinanceToNotionUseCase{
		e:   e,
		v:   v,
		nac: nac,
		mac: mac,
	}
}

type OpenFinanceToNotionUseCaseDTO struct {
	StartDate string `validate:"datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `validate:"datetime=2006-01-02T15:04:05Z07:00"`
}

func (uc *OpenFinanceToNotionUseCase) Execute(
	dto OpenFinanceToNotionUseCaseDTO,
) error {
	uc.setDefaultValues(&dto)

	if err := uc.v.Validate(dto); err != nil {
		return err
	}

	startDate, endDate, err := uc.parseDates(dto)
	if err != nil {
		return err
	}

	if err := uc.mac.Authenticate(); err != nil {
		return fmt.Errorf("error authenticating meupluggyapi client: %w", err)
	}

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := []error{}

	meupluggyapiResults := []meupluggyapi.Result{}

	wg.Add(len(uc.e.MeuPluggyAccountIDs))

	for _, accountID := range uc.e.MeuPluggyAccountIDs {
		go func() {
			defer wg.Done()

			res, err := uc.mac.ListTransactions(
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
			meupluggyapiResults = append(meupluggyapiResults, res.Results...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error fetching transactions:\n%v", errs)
	}

	transactions := []entity.Transaction{}
	uniqueCategories := map[string]struct{}{}

	for _, r := range meupluggyapiResults {
		if isInvestment := (r.Category != nil && *r.Category == "Investments") ||
			(r.Description == "Aplicação RDB"); isInvestment {
			continue
		}
		if isReceivingMoney := r.Type == meupluggyapi.Credit; isReceivingMoney {
			continue
		}
		if isCreditCardBillPayment := r.
			Description == "Pagamento de fatura"; isCreditCardBillPayment {
			continue
		}

		transaction := entity.Transaction{
			Amount: math.Abs(r.Amount),
			Date:   r.Date,
		}

		if r.Category != nil {
			uniqueCategories[*r.Category] = struct{}{}
			transaction.Category = *r.Category
		}

		accountType := entity.CreditCard
		if r.PaymentData != nil {
			accountType = entity.BankAccount
		}

		if accountType == entity.BankAccount {
			transaction.PaymentMethod = *r.PaymentData.PaymentMethod
			transaction.Description = r.Description

			if hasReceiver := (r.PaymentData.Receiver != nil); hasReceiver {
				if hasReceiverName := r.PaymentData.
					Receiver.Name != nil; hasReceiverName {
					transaction.Name = *r.PaymentData.Receiver.Name
				} else if hasReceiverDocument := r.PaymentData.
					Receiver.DocumentNumber != nil; hasReceiverDocument {
					document, _ := formatter.MaskDocument(
						r.PaymentData.Receiver.DocumentNumber.Value,
						r.PaymentData.Receiver.DocumentNumber.Type,
					)
					transaction.Name = document
				}
			}
		} else if accountType == entity.CreditCard {
			transaction.Name = r.Description
			transaction.PaymentMethod = "CREDIT CARD"
		}

		transactions = append(transactions, transaction)
	}

	categories := []string{}
	for category := range uniqueCategories {
		categories = append(categories, category)
	}

	year, month, _ := startDate.Date()
	strMonth := dateutil.MonthMapper[month]
	title := fmt.Sprintf("%s %d", strMonth, year)

	newTableResponse, err := uc.nac.NewTable(sheet.NewTableDTO{
		ParentID:   uc.e.NotionPageID,
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
			if _, err := uc.nac.
				InsertRow(newTableResponse.ID, transaction); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error fetching transactions:\n%v", errs)
	}

	return nil
}

func (uc *OpenFinanceToNotionUseCase) setDefaultValues(
	dto *OpenFinanceToNotionUseCaseDTO,
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

func (uc *OpenFinanceToNotionUseCase) parseDates(
	dto OpenFinanceToNotionUseCaseDTO,
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
