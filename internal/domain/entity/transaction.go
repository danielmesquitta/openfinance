package entity

import (
	"fmt"
	"time"
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
		"%s:%f:%s",
		t.Name,
		t.Amount,
		t.Date.Format(time.DateTime),
	)
}
