package pluggyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
	"golang.org/x/sync/errgroup"
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

const resultLogField = "result"

func (c *Client) ListTransactionsByUserID(
	ctx context.Context,
	userID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	resTransactions, err := c.fetchAllAccountTransactions(
		ctx, conn.accountIDs, conn.accessToken, from, to)
	if err != nil {
		return nil, err
	}

	transactions := []entity.Transaction{}
	for _, data := range resTransactions {
		t := c.parseRequestToTransactions(data)
		transactions = append(transactions, t...)
	}

	return transactions, nil
}

func (c *Client) fetchAccountTransactions(
	ctx context.Context,
	accountID, accessToken string,
	from, to time.Time,
) (listTransactionsResponse, error) {
	res, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"pageSize":  "500",
			"from":      from.Format(time.DateOnly),
			"to":        to.Format(time.DateOnly),
			"accountId": accountID,
		}).
		SetHeader("X-API-KEY", accessToken).
		Get("/transactions")
	if err != nil {
		return listTransactionsResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	body := res.Body()
	if res.IsError() {
		return listTransactionsResponse{}, fmt.Errorf(
			"error response while listing transactions: %s", body)
	}

	data := listTransactionsResponse{}
	if err := json.Unmarshal(body, &data); err != nil {
		return listTransactionsResponse{}, fmt.Errorf(
			"failed to unmarshal while listing transactions: %w", err)
	}

	return data, nil
}

func (c *Client) fetchAllAccountTransactions(
	ctx context.Context,
	accountIDs []string,
	accessToken string,
	from, to time.Time,
) ([]listTransactionsResponse, error) {
	resTransactions := make([]listTransactionsResponse, len(accountIDs))
	g, gCtx := errgroup.WithContext(ctx)

	for i, accountID := range accountIDs {
		g.Go(func() error {
			data, err := c.fetchAccountTransactions(gCtx, accountID, accessToken, from, to)
			if err != nil {
				return err
			}
			resTransactions[i] = data
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to wait for listing transactions: %w", err)
	}

	return resTransactions, nil
}

func (c *Client) shouldSkipTransaction(r result) bool {
	if isInvestment := (r.Category != nil && *r.Category == "Investments") ||
		strings.Contains(r.Description, "Aplicação"); isInvestment {
		return true
	}
	if isReceivingMoney := r.Type == Credit; isReceivingMoney {
		return true
	}
	if isCreditCardBillPayment := r.
		Description == "Pagamento de fatura"; isCreditCardBillPayment {
		return true
	}
	return false
}

func (c *Client) setTransactionNameFromReceiver(transaction *entity.Transaction, r result) bool {
	if hasReceiver := (r.PaymentData.Receiver != nil); !hasReceiver {
		return false
	}

	if hasReceiverName := r.PaymentData.Receiver.Name != nil; hasReceiverName {
		transaction.Name = *r.PaymentData.Receiver.Name
		return false
	}

	if hasReceiverDocument := r.PaymentData.Receiver.DocumentNumber != nil; hasReceiverDocument {
		document, err := docutil.MaskDocument(r.PaymentData.Receiver.DocumentNumber.Value)
		if err != nil {
			slog.Error("error masking document", "error", err)
			return true
		}
		transaction.Name = document
		return false
	}

	return false
}

func (c *Client) handleBankTransaction(transaction *entity.Transaction, r result) bool {
	if r.PaymentData == nil {
		slog.Error("PaymentData is nil", resultLogField, r)
		return true
	}

	if r.PaymentData.PaymentMethod == nil {
		slog.Error("PaymentMethod is nil", resultLogField, r)
		return true
	}

	transaction.PaymentMethod = *r.PaymentData.PaymentMethod

	if r.Description != "" {
		transaction.Name = r.Description
		return false
	}

	return c.setTransactionNameFromReceiver(transaction, r)
}

func (c *Client) processSingleResult(r result) (*entity.Transaction, bool) {
	transaction := entity.Transaction{
		Amount: math.Abs(r.Amount),
		Date:   r.Date,
	}

	if r.Category != nil {
		transaction.Category = entity.Category(*r.Category)
	}

	accountType := entity.AccountTypeCreditCard
	if r.PaymentData != nil {
		accountType = entity.AccountTypeBank
	}

	switch accountType {
	case entity.AccountTypeBank:
		if shouldContinue := c.handleBankTransaction(&transaction, r); shouldContinue {
			return nil, true
		}
	case entity.AccountTypeCreditCard:
		transaction.Name = r.Description
		transaction.PaymentMethod = entity.PaymentMethodCreditCard
	default:
		slog.Error("unknown account type", "accountType", accountType, resultLogField, r)
		return nil, true
	}

	return &transaction, false
}

func (c *Client) parseRequestToTransactions(
	data listTransactionsResponse,
) []entity.Transaction {
	transactions := []entity.Transaction{}

	for _, r := range data.Results {
		if c.shouldSkipTransaction(r) {
			continue
		}

		transaction, shouldSkip := c.processSingleResult(r)
		if shouldSkip {
			continue
		}

		transactions = append(transactions, *transaction)
	}

	return transactions
}
