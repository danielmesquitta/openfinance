package pluggyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

const transactionPageSize = "500"

type listTransactionsResponse struct {
	TotalPages int64                            `json:"totalPages"`
	Page       int64                            `json:"page"`
	Results    []listTransactionsResponseResult `json:"results"`
}

type listTransactionsResponseResult struct {
	Description             string              `json:"description"`
	Amount                  float64             `json:"amount"`
	AmountInAccountCurrency *float64            `json:"amountInAccountCurrency"`
	Date                    time.Time           `json:"date"`
	Category                *string             `json:"category"`
	PaymentData             *paymentData        `json:"paymentData"`
	Type                    string              `json:"type"`
	CreditCardMetadata      *creditCardMetadata `json:"creditCardMetadata"`
}

type creditCardMetadata struct {
	CardNumber *string `json:"cardNumber,omitempty"`
}

type paymentData struct {
	PaymentMethod *entity.PaymentMethod `json:"paymentMethod"`
	Receiver      *payer                `json:"receiver"`
}

type payer struct {
	Name           *string         `json:"name"`
	DocumentNumber *documentNumber `json:"documentNumber"`
}

type documentNumber struct {
	Value string `json:"value"`
}

func (c *Client) ListTransactionsByIngestProfileID(
	ctx context.Context,
	ingestProfileID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	connection, ok := c.conns[ingestProfileID]
	if !ok {
		return nil, errors.New("connection not found for ingest profile " + ingestProfileID)
	}

	results, err := c.fetchAllAccountTransactions(
		ctx,
		connection.accountIDs,
		connection.accessToken,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	transactions := make([]entity.Transaction, 0, len(results))
	for _, result := range results {
		transaction, accepted := entity.NewTransaction(transactionInput(result))
		if accepted {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}

func (c *Client) fetchAccountTransactionsPage(
	ctx context.Context,
	accountID, accessToken string,
	from, to time.Time,
	page int,
) (listTransactionsResponse, error) {
	response, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"pageSize":  transactionPageSize,
			"page":      strconv.Itoa(page),
			"from":      from.Format(time.DateOnly),
			"to":        to.Format(time.DateOnly),
			"accountId": accountID,
		}).
		SetHeader("X-API-KEY", accessToken).
		Get("/transactions")
	if err != nil {
		return listTransactionsResponse{}, fmt.Errorf("list transactions: %w", err)
	}

	if response.IsError() {
		return listTransactionsResponse{}, fmt.Errorf(
			"list transactions for account %s from %s to %s: %s",
			accountID,
			from.Format(time.DateOnly),
			to.Format(time.DateOnly),
			response.Body(),
		)
	}

	data := listTransactionsResponse{}
	if err := json.Unmarshal(response.Body(), &data); err != nil {
		return listTransactionsResponse{}, fmt.Errorf("decode transactions response: %w", err)
	}

	return data, nil
}

func (c *Client) fetchAccountTransactions(
	ctx context.Context,
	accountID, accessToken string,
	from, to time.Time,
) ([]listTransactionsResponseResult, error) {
	var results []listTransactionsResponseResult
	for page := 1; ; page++ {
		data, err := c.fetchAccountTransactionsPage(ctx, accountID, accessToken, from, to, page)
		if err != nil {
			return nil, err
		}

		results = append(results, data.Results...)
		if data.TotalPages <= int64(page) {
			break
		}
	}

	return results, nil
}

func (c *Client) fetchAllAccountTransactions(
	ctx context.Context,
	accountIDs []string,
	accessToken string,
	from, to time.Time,
) ([]listTransactionsResponseResult, error) {
	resultsByAccount := make([][]listTransactionsResponseResult, len(accountIDs))
	group, groupContext := errgroup.WithContext(ctx)
	group.SetLimit(c.maxConcurrentOperations)

	for index, accountID := range accountIDs {
		group.Go(func() error {
			select {
			case c.accountSlots <- struct{}{}:
				defer func() { <-c.accountSlots }()
			case <-groupContext.Done():
				return groupContext.Err()
			}

			results, err := c.fetchAccountTransactions(groupContext, accountID, accessToken, from, to)
			if err != nil {
				return fmt.Errorf("list account %s transactions: %w", accountID, err)
			}

			resultsByAccount[index] = results

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, fmt.Errorf("wait for account transactions: %w", err)
	}

	var results []listTransactionsResponseResult
	for _, accountResults := range resultsByAccount {
		results = append(results, accountResults...)
	}

	return results, nil
}

func transactionInput(result listTransactionsResponseResult) entity.TransactionInput {
	input := entity.TransactionInput{
		AccountType:             entity.AccountTypeCreditCard,
		Description:             result.Description,
		Amount:                  result.Amount,
		AmountInAccountCurrency: result.AmountInAccountCurrency,
		Date:                    result.Date,
		Direction:               entity.TransactionDirection(result.Type),
	}

	if result.Category != nil {
		input.SourceCategory = *result.Category
	}

	if result.CreditCardMetadata != nil {
		input.CardLastDigits = result.CreditCardMetadata.CardNumber
	}

	if result.PaymentData == nil {
		return input
	}

	input.AccountType = entity.AccountTypeBank
	input.PaymentMethod = result.PaymentData.PaymentMethod
	if result.PaymentData.Receiver == nil {
		return input
	}

	if result.PaymentData.Receiver.Name != nil {
		input.ReceiverName = *result.PaymentData.Receiver.Name
	}
	if result.PaymentData.Receiver.DocumentNumber != nil {
		input.ReceiverDocument = result.PaymentData.Receiver.DocumentNumber.Value
	}

	return input
}
