package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type ListTransactionsResponse struct {
	Total      int64    `json:"total"`
	TotalPages int64    `json:"totalPages"`
	Page       int64    `json:"page"`
	Results    []Result `json:"results"`
}

type Result struct {
	ID                      string              `json:"id"`
	Description             string              `json:"description"`
	Amount                  float64             `json:"amount"`
	AmountInAccountCurrency *float64            `json:"amountInAccountCurrency"`
	Date                    time.Time           `json:"date"`
	Category                *string             `json:"category"`
	PaymentData             *PaymentData        `json:"paymentData"`
	Type                    ResultType          `json:"type"`
	CreditCardMetadata      *CreditCardMetadata `json:"creditCardMetadata"`
}

type CreditCardMetadata struct {
	CardNumber        *string `json:"cardNumber,omitempty"`
	TotalInstallments *int64  `json:"totalInstallments,omitempty"`
	InstallmentNumber *int64  `json:"installmentNumber,omitempty"`
}

type PaymentData struct {
	Payer         *Payer                `json:"payer"`
	PaymentMethod *entity.PaymentMethod `json:"paymentMethod"`
	Receiver      *Payer                `json:"receiver"`
}

type Payer struct {
	Name           *string         `json:"name"`
	DocumentNumber *DocumentNumber `json:"documentNumber"`
}

type DocumentNumber struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ResultType string

const (
	Credit ResultType = "CREDIT"
	Debit  ResultType = "DEBIT"
)

func (c *Client) ListTransactions(
	accountID string,
	from, to time.Time,
) (*ListTransactionsResponse, error) {
	url := c.BaseURL

	url.Path = "/transactions"
	query := url.Query()

	query.Add("accountId", accountID)
	query.Add("pageSize", "500")

	query.Add("from", (from).Format(time.DateOnly))
	query.Add("to", (to).Format(time.DateOnly))

	req, err := http.NewRequest("GET", url.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", c.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := &ListTransactionsResponse{}
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return data, nil
}
