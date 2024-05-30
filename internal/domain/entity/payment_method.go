package entity

type PaymentMethod string

const (
	Boleto PaymentMethod = "BOLETO"
	Pix    PaymentMethod = "PIX"
	Ted    PaymentMethod = "TED"
)
