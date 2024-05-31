package entity

import "time"

type Transaction struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Category      string        `json:"category"`
	Amount        float64       `json:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Date          time.Time     `json:"date"`
}
