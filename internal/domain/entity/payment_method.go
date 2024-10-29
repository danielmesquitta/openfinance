package entity

type PaymentMethod string

const (
	PaymentMethodBoleto     PaymentMethod = "BOLETO"
	PaymentMethodPix        PaymentMethod = "PIX"
	PaymentMethodTed        PaymentMethod = "TED"
	PaymentMethodCreditCard PaymentMethod = "CREDIT CARD"
)
