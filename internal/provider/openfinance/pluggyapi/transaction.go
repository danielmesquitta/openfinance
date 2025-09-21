package pluggyapi

import (
	"cmp"
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
	"github.com/danielmesquitta/openfinance/internal/pkg/ptr"
	"golang.org/x/sync/errgroup"
)

type listTransactionsResp struct {
	Total      int64                        `json:"total"`
	TotalPages int64                        `json:"totalPages"`
	Page       int64                        `json:"page"`
	Results    []listTransactionsRespResult `json:"results"`
}

type listTransactionsRespResult struct {
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

const (
	samePersonTransfer = "Same person transfer"
	resultLogField     = "listTransactionsRespResult"
)

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
) (listTransactionsResp, error) {
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
		return listTransactionsResp{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	body := res.Body()
	if res.IsError() {
		return listTransactionsResp{}, fmt.Errorf(
			"failed to list transactions with account id %s and date range %s to %s: %s",
			accountID,
			from,
			to,
			body,
		)
	}

	data := listTransactionsResp{}
	if err := json.Unmarshal(body, &data); err != nil {
		return listTransactionsResp{}, fmt.Errorf(
			"failed to unmarshal while listing transactions with account id %s and date range %s to %s: %w",
			accountID,
			from,
			to,
			err,
		)
	}

	return data, nil
}

func (c *Client) fetchAllAccountTransactions(
	ctx context.Context,
	accountIDs []string,
	accessToken string,
	from, to time.Time,
) ([]listTransactionsResp, error) {
	resTransactions := make([]listTransactionsResp, len(accountIDs))
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

func (c *Client) shouldSkipTransaction(r listTransactionsRespResult) bool {
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

	if r.Category != nil && *r.Category == samePersonTransfer {
		return true
	}

	return false
}

func (c *Client) setTransactionNameFromReceiver(
	transaction *entity.Transaction,
	r listTransactionsRespResult,
) bool {
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

func (c *Client) handleBankTransaction(
	transaction *entity.Transaction,
	r listTransactionsRespResult,
) bool {
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

func (c *Client) processSingleResult(r listTransactionsRespResult) (*entity.Transaction, bool) {
	amountInAccountCurrency := ptr.Deref(r.AmountInAccountCurrency)
	amount := math.Abs(cmp.Or(amountInAccountCurrency, r.Amount))

	transaction := entity.Transaction{
		Amount: amount,
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

		if r.CreditCardMetadata != nil && r.CreditCardMetadata.CardNumber != nil {
			transaction.CardLastDigits = r.CreditCardMetadata.CardNumber
		}

	default:
		slog.Error("unknown account type", "accountType", accountType, resultLogField, r)

		return nil, true
	}

	return &transaction, false
}

func (c *Client) parseRequestToTransactions(
	data listTransactionsResp,
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
