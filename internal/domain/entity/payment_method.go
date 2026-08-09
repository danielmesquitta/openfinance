package entity

type PaymentMethod string

const (
	PaymentMethodBoleto     PaymentMethod = "BOLETO"
	PaymentMethodPix        PaymentMethod = "PIX"
	PaymentMethodTed        PaymentMethod = "TED"
	PaymentMethodCreditCard PaymentMethod = "CREDIT CARD"
)

var PaymentMethods = []PaymentMethod{
	PaymentMethodBoleto,
	PaymentMethodPix,
	PaymentMethodTed,
	PaymentMethodCreditCard,
}

var PaymentMethodColors = map[PaymentMethod]Color{
	PaymentMethodBoleto:     Yellow,
	PaymentMethodPix:        Blue,
	PaymentMethodTed:        Green,
	PaymentMethodCreditCard: Purple,
}
