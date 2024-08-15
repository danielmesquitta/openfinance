package meupluggyapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
)

type _listTransactionsResponse struct {
	Total      int64     `json:"total"`
	TotalPages int64     `json:"totalPages"`
	Page       int64     `json:"page"`
	Results    []_result `json:"results"`
}

type _result struct {
	ID                      string               `json:"id"`
	Description             string               `json:"description"`
	Amount                  float64              `json:"amount"`
	AmountInAccountCurrency *float64             `json:"amountInAccountCurrency"`
	Date                    time.Time            `json:"date"`
	Category                *string              `json:"category"`
	PaymentData             *_paymentData        `json:"paymentData"`
	Type                    _resultType          `json:"type"`
	CreditCardMetadata      *_creditCardMetadata `json:"creditCardMetadata"`
}

type _creditCardMetadata struct {
	CardNumber        *string `json:"cardNumber,omitempty"`
	TotalInstallments *int64  `json:"totalInstallments,omitempty"`
	InstallmentNumber *int64  `json:"installmentNumber,omitempty"`
}

type _paymentData struct {
	Payer         *_payer               `json:"payer"`
	PaymentMethod *entity.PaymentMethod `json:"paymentMethod"`
	Receiver      *_payer               `json:"receiver"`
}

type _payer struct {
	Name           *string          `json:"name"`
	DocumentNumber *_documentNumber `json:"documentNumber"`
}

type _documentNumber struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type _resultType string

const (
	Credit _resultType = "CREDIT"
	Debit  _resultType = "DEBIT"
)

func (c *Client) ListTransactionsByUserID(
	userID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	conn := c.conns[userID]

	jobsCount := len(conn.accountIDs)
	ch := make(chan _listTransactionsResponse, jobsCount)
	errCh := make(chan error, jobsCount)
	wg := sync.WaitGroup{}
	wg.Add(jobsCount)

	for _, accountID := range conn.accountIDs {
		go func() {
			defer wg.Done()

			url := c.baseURL
			url.Path = "/transactions"
			query := url.Query()

			query.Add("pageSize", "500")
			query.Add("from", from.Format(time.DateOnly))
			query.Add("to", to.Format(time.DateOnly))

			query.Add("accountId", accountID)

			fullURL := url.String() + "?" + query.Encode()

			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				errCh <- entity.NewErr(err)
				return
			}

			req.Header.Add("accept", "application/json")
			req.Header.Add("X-API-KEY", conn.accessToken)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				errCh <- entity.NewErr(err)
				return
			}
			if res == nil {
				errCh <- entity.NewErr("response is nil")
				return
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				errCh <- parseResError(res)
				return
			}

			decoder := json.NewDecoder(res.Body)
			data := _listTransactionsResponse{}
			if err := decoder.Decode(&data); err != nil {
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
	data _listTransactionsResponse,
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

		accountType := entity.CreditCard
		if r.PaymentData != nil {
			accountType = entity.BankAccount
		}

		switch accountType {
		case entity.BankAccount:
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

		case entity.CreditCard:
			transaction.Name = r.Description
			transaction.PaymentMethod = "CREDIT CARD"
		}

	appendTransaction:
		transactions = append(transactions, transaction)
	}

	return transactions
}
