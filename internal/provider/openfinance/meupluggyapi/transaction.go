package meupluggyapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
)

type listTransactionsResponse struct {
	Total      int64    `json:"total"`
	TotalPages int64    `json:"totalPages"`
	Page       int64    `json:"page"`
	Results    []result `json:"results"`
}

type result struct {
	ID                      string              `json:"id"`
	Description             string              `json:"description"`
	Amount                  float64             `json:"amount"`
	AmountInAccountCurrency *float64            `json:"amountInAccountCurrency"`
	Date                    time.Time           `json:"date"`
	Category                *string             `json:"category"`
	PaymentData             *paymentData        `json:"paymentData"`
	Type                    resultType          `json:"type"`
	CreditCardMetadata      *creditCardMetadata `json:"creditCardMetadata"`
}

type creditCardMetadata struct {
	CardNumber        *string `json:"cardNumber,omitempty"`
	TotalInstallments *int64  `json:"totalInstallments,omitempty"`
	InstallmentNumber *int64  `json:"installmentNumber,omitempty"`
}

type paymentData struct {
	Payer         *payer                `json:"payer"`
	PaymentMethod *entity.PaymentMethod `json:"paymentMethod"`
	Receiver      *payer                `json:"receiver"`
}

type payer struct {
	Name           *string         `json:"name"`
	DocumentNumber *documentNumber `json:"documentNumber"`
}

type documentNumber struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type resultType string

const (
	Credit resultType = "CREDIT"
	Debit  resultType = "DEBIT"
)

func (c *Client) ListTransactionsByUserID(
	userID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, entity.NewErr("connection not found for user " + userID)
	}

	jobsCount := len(conn.accountIDs)
	ch := make(chan listTransactionsResponse, jobsCount)
	errCh := make(chan error, jobsCount)
	wg := sync.WaitGroup{}
	wg.Add(jobsCount)

	for _, accountID := range conn.accountIDs {
		go func() {
			defer wg.Done()
			res, err := c.client.R().
				SetQueryParams(map[string]string{
					"pageSize":  "500",
					"from":      from.Format(time.DateOnly),
					"to":        to.Format(time.DateOnly),
					"accountId": accountID,
				}).
				SetHeader("X-API-KEY", conn.accessToken).
				Get("/transactions")
			if err != nil {
				errCh <- entity.NewErr(err)
				return
			}
			body := res.Body()
			if statusCode := res.StatusCode(); statusCode < 200 ||
				statusCode >= 300 {
				errCh <- entity.NewErr(body)
				return
			}

			data := listTransactionsResponse{}
			if err := json.Unmarshal(body, &data); err != nil {
				errCh <- entity.NewErr(err)
				return
			}

			ch <- data
		}()
	}

	wg.Wait()
	close(errCh)
	close(ch)

	var err error
	for e := range errCh {
		err = errors.Join(err, e)
	}
	if err != nil {
		return nil, entity.NewErr(err)
	}

	transactions := []entity.Transaction{}
	for data := range ch {
		t := c.parseRequestToTransactions(data)
		transactions = append(transactions, t...)
	}

	return transactions, nil
}

func (c *Client) parseRequestToTransactions(
	data listTransactionsResponse,
) []entity.Transaction {
	transactions := []entity.Transaction{}

loop:
	for _, r := range data.Results {
		if isInvestment := (r.Category != nil && *r.Category == "Investments") ||
			strings.Contains(r.Description, "Aplicação"); isInvestment {
			continue
		}
		if isReceivingMoney := r.Type == Credit; isReceivingMoney {
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
			transaction.Category = *r.Category
		}

		accountType := entity.AccountTypeCreditCard
		if r.PaymentData != nil {
			accountType = entity.AccountTypeBank
		}

		switch accountType {
		case entity.AccountTypeBank:
			if r.PaymentData == nil {
				slog.Error("PaymentData is nil", "result", r)
				continue loop
			}

			if r.PaymentData.PaymentMethod == nil {
				slog.Error("PaymentMethod is nil", "result", r)
				continue loop
			}

			transaction.PaymentMethod = *r.PaymentData.PaymentMethod
			transaction.Description = r.Description

			if hasReceiver := (r.PaymentData.Receiver != nil); !hasReceiver {
				goto appendTransaction
			}

			if hasReceiverName := r.PaymentData.
				Receiver.Name != nil; hasReceiverName {
				transaction.Name = *r.PaymentData.Receiver.Name
				goto appendTransaction
			}

			if hasReceiverDocument := r.PaymentData.
				Receiver.DocumentNumber != nil; hasReceiverDocument {
				document, err := docutil.MaskDocument(r.PaymentData.Receiver.DocumentNumber.Value)
				if err != nil {
					slog.Error("error masking document", "error", err)
					continue loop
				}

				transaction.Name = document
				goto appendTransaction
			}

		case entity.AccountTypeCreditCard:
			transaction.Name = r.Description
			transaction.PaymentMethod = entity.PaymentMethodCreditCard
		}

	appendTransaction:
		transactions = append(transactions, transaction)
	}

	return transactions
}
