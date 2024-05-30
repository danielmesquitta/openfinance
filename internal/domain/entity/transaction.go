package entity

import "time"

type Transaction struct {
	Name          string
	Description   string
	Category      string
	Amount        float64
	PaymentMethod PaymentMethod
	Date          time.Time
}
