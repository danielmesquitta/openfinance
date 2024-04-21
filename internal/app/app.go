package app

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

func Run() error {
	e := config.LoadEnv()

	meupluggyAPIClient := meupluggyapi.NewClient(
		e.MeuPluggyClientID,
		e.MeuPluggyClientSecret,
	)

	if err := meupluggyAPIClient.Authenticate(); err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}

	now := time.Now()
	startOfMonth := time.Date(
		now.Year(),
		now.Month(),
		1,
		0,
		0,
		0,
		0,
		now.Location(),
	)

	mu := sync.Mutex{}
	errs := []error{}

	requestsData := []string{
		e.MeuPluggyBankAccountID,
		e.MeuPluggyCreditAccountID,
	}

	uniqueCategories := map[string]struct{}{}
	transactions := []notionapi.InsertRowDTO{}

	asyncloop.Loop(requestsData, func(i int, v string) {
		res, err := meupluggyAPIClient.ListTransactions(
			v,
			&startOfMonth,
			nil,
		)
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
		}

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
					continue
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
							continue
						}

						transaction.Name = document
					}
				}

			case CreditCard:
				transaction.Name = r.Description
				transaction.PaymentMethod = "CREDIT CARD"
			}

			transactions = append(transactions, transaction)
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

	notionAPIClient := notionapi.NewClient(e.NotionToken)

	createDBRes, err := notionAPIClient.CreateDB(notionapi.CreateDBDTO{
		PageID:     e.NotionPageID,
		Date:       startOfMonth,
		Categories: categories,
	})
	if err != nil {
		return fmt.Errorf("error creating spending database: %v", err)
	}

	asyncloop.Loop(transactions, func(_ int, t notionapi.InsertRowDTO) {
		t.DatabaseID = createDBRes.ID
		if _, err := notionAPIClient.InsertRow(t); err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return
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
