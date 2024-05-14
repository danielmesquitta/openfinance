package usecase

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/danielmesquitta/asyncloop"

	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/service/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/formatter"
)

type AccountType int

const (
	BankAccount AccountType = iota
	CreditCard
)

type OpenFinanceToNotionUseCase struct {
	env                *config.Env
	notionAPIClient    *notionapi.Client
	meupluggyAPIClient *meupluggyapi.Client
}

func NewOpenFinanceToNotionUseCase(
	env *config.Env,
	notionAPIClient *notionapi.Client,
	meupluggyAPIClient *meupluggyapi.Client,
) *OpenFinanceToNotionUseCase {
	return &OpenFinanceToNotionUseCase{
		env:                env,
		notionAPIClient:    notionAPIClient,
		meupluggyAPIClient: meupluggyAPIClient,
	}
}

func (uc *OpenFinanceToNotionUseCase) Execute() error {
	if err := uc.meupluggyAPIClient.Authenticate(); err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	now := time.Now()
	startOfMonth := time.Date(
		2024,
		5,
		1,
		0,
		0,
		0,
		0,
		now.Location(),
	)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	mu := sync.Mutex{}
	errs := []error{}

	transactionMu := sync.Mutex{}
	transactions := []notionapi.InsertRowDTO{}

	uniqueCategories := map[string]struct{}{}

	asyncloop.Loop(uc.env.MeuPluggyAccountIDs, func(i int, v string) {
		res, err := uc.meupluggyAPIClient.ListTransactions(
			v,
			&startOfMonth,
			&endOfMonth,
		)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}

	loop:
		for _, r := range res.Results {
			if isInvestment := (r.Category != nil &&
				*r.Category == "Investments"); isInvestment {
				continue
			}
			if isReceivingMoney := r.Type == meupluggyapi.Credit; isReceivingMoney {
				continue
			}
			if isCreditCardBillPayment := r.
				Description == "Pagamento de fatura"; isCreditCardBillPayment {
				continue
			}

			transaction := notionapi.InsertRowDTO{
				Amount: math.Abs(r.Amount),
				Date:   r.Date,
			}

			if r.Category != nil {
				uniqueCategories[*r.Category] = struct{}{}
				transaction.Category = *r.Category
			}

			switch AccountType(i) {
			case BankAccount:
				if r.PaymentData == nil {
					mu.Lock()
					errs = append(
						errs,
						fmt.Errorf(
							"PaymentData in transaction %v is nil",
							r.ID,
						),
					)
					mu.Unlock()
					continue loop
				}

				transaction.PaymentMethod = *r.PaymentData.PaymentMethod
				transaction.Description = r.Description

				if hasReceiver := (r.PaymentData.Receiver != nil); hasReceiver {
					if hasReceiverName := r.PaymentData.
						Receiver.Name != nil; hasReceiverName {
						transaction.Name = *r.PaymentData.Receiver.Name
					} else if hasReceiverDocument := r.PaymentData.
						Receiver.DocumentNumber != nil; hasReceiverDocument {
						document, err := formatter.MaskDocument(
							r.PaymentData.Receiver.DocumentNumber.Value,
							r.PaymentData.Receiver.DocumentNumber.Type,
						)

						if err != nil {
							mu.Lock()
							errs = append(errs, err)
							mu.Unlock()
							continue loop
						}

						transaction.Name = document
					}
				}

			case CreditCard:
				transaction.Name = r.Description
				transaction.PaymentMethod = "CREDIT CARD"
			}

			transactionMu.Lock()
			transactions = append(transactions, transaction)
			transactionMu.Unlock()
		}
	})

	if len(errs) > 0 {
		errMsgs := []string{}
		for _, err := range errs {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf(
			"error fetching transactions:\n%v",
			strings.Join(errMsgs, "\n\n"),
		)
	}

	categories := []string{}
	for category := range uniqueCategories {
		categories = append(categories, category)
	}

	createDBRes, err := uc.notionAPIClient.CreateDB(notionapi.CreateDBDTO{
		PageID:     uc.env.NotionPageID,
		Date:       startOfMonth,
		Categories: categories,
	})
	if err != nil {
		return fmt.Errorf("error creating spending database: %v", err)
	}

	asyncloop.Loop(transactions, func(i int, t notionapi.InsertRowDTO) {
		t.DatabaseID = createDBRes.ID
		if _, err := uc.notionAPIClient.InsertRow(t); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	})

	if len(errs) > 0 {
		errMsgs := []string{}
		for _, err := range errs {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf(
			"error fetching transactions:\n%v",
			strings.Join(errMsgs, "\n\n"),
		)
	}

	return nil
}
