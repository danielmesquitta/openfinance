package entity

import (
	"fmt"
	"math"
	"time"
)

const (
	dateTimeWithoutSeconds = "2006-01-02 15:04"
)

type Transaction struct {
	Name           string        `json:"name,omitzero"`
	Category       Category      `json:"category,omitzero"`
	Amount         float64       `json:"amount,omitzero"`
	PaymentMethod  PaymentMethod `json:"payment_method,omitzero"`
	Date           time.Time     `json:"date,omitzero"`
	CardLastDigits *string       `json:"card_last_digits,omitzero"`
}

func (t *Transaction) ID() string {
	return fmt.Sprintf(
		"%s:%d:%s",
		t.Name,
		int64(math.Round(t.Amount*100)), // multiply by 100 to avoid floating point precision issues
		t.Date.Format(dateTimeWithoutSeconds),
	)
}
