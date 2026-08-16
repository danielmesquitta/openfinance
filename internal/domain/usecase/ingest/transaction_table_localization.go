package ingest

import (
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

const monthsPerYear = 12

type transactionTableColumns struct {
	name           string
	category       string
	budgetGroup    string
	amount         string
	paymentMethod  string
	cardLastDigits string
	date           string
}

type transactionTableLocalization struct {
	columns               transactionTableColumns
	paymentMethodLabels   map[entity.PaymentMethod]string
	paymentMethodsByLabel map[string]entity.PaymentMethod
	monthNames            [monthsPerYear]string
}

var transactionTableLocalizations = map[entity.Language]transactionTableLocalization{
	entity.LanguageEnglish: newTransactionTableLocalization(
		transactionTableColumns{
			name:           "Name",
			category:       "Category",
			budgetGroup:    "Budget Group",
			amount:         "Amount",
			paymentMethod:  "Payment Method",
			cardLastDigits: "Card Last Digits",
			date:           "Date",
		},
		map[entity.PaymentMethod]string{
			entity.PaymentMethodBoleto:     "BOLETO",
			entity.PaymentMethodPix:        "PIX",
			entity.PaymentMethodTed:        "TED",
			entity.PaymentMethodCreditCard: "CREDIT CARD",
		},
		[monthsPerYear]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	),
	entity.LanguagePortugueseBrazil: newTransactionTableLocalization(
		transactionTableColumns{
			name:           "Nome",
			category:       "Categoria",
			budgetGroup:    "Grupo do orçamento",
			amount:         "Valor",
			paymentMethod:  "Forma de pagamento",
			cardLastDigits: "Últimos dígitos do cartão",
			date:           "Data",
		},
		map[entity.PaymentMethod]string{
			entity.PaymentMethodBoleto:     "BOLETO",
			entity.PaymentMethodPix:        "PIX",
			entity.PaymentMethodTed:        "TED",
			entity.PaymentMethodCreditCard: "CARTÃO DE CRÉDITO",
		},
		[monthsPerYear]string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"},
	),
}

func newTransactionTableLocalization(
	columns transactionTableColumns,
	paymentMethodLabels map[entity.PaymentMethod]string,
	monthNames [monthsPerYear]string,
) transactionTableLocalization {
	paymentMethodsByLabel := make(map[string]entity.PaymentMethod, len(paymentMethodLabels))
	for paymentMethod, label := range paymentMethodLabels {
		paymentMethodsByLabel[label] = paymentMethod
	}

	return transactionTableLocalization{
		columns:               columns,
		paymentMethodLabels:   paymentMethodLabels,
		paymentMethodsByLabel: paymentMethodsByLabel,
		monthNames:            monthNames,
	}
}

func normalizedLanguage(language entity.Language) entity.Language {
	if language == "" {
		return entity.DefaultLanguage
	}

	return language
}

func transactionTableLocalizationFor(language entity.Language) transactionTableLocalization {
	language = normalizedLanguage(language)
	localization, exists := transactionTableLocalizations[language]
	if !exists {
		return transactionTableLocalizations[entity.DefaultLanguage]
	}

	return localization
}

func localizedTransactionTableTitle(month time.Time, language entity.Language) string {
	localization := transactionTableLocalizationFor(language)

	return fmt.Sprintf("%s %d", localization.monthNames[month.Month()-1], month.Year())
}

func alternateLanguage(language entity.Language) entity.Language {
	if normalizedLanguage(language) == entity.LanguagePortugueseBrazil {
		return entity.LanguageEnglish
	}

	return entity.LanguagePortugueseBrazil
}
