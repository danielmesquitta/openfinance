package ingest

import (
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

const (
	transactionTableIcon = "💸"
	transactionCurrency  = "BRL"
)

func transactionTableOptions(settings entity.IngestProfileSettings) []sheet.CreateTableOption {
	localization := transactionTableLocalizationFor(settings.Language)
	columns := localization.columns

	categoryOptions := make([]sheet.SelectOption, 0, len(settings.Categories))
	for _, category := range settings.Categories {
		categoryOptions = append(
			categoryOptions,
			sheet.NewSelectOption(
				string(category),
				sheet.WithColor(settings.ColorsByCategory[category]),
			),
		)
	}

	paymentMethodOptions := make([]sheet.SelectOption, 0, len(entity.PaymentMethods))
	for _, paymentMethod := range entity.PaymentMethods {
		paymentMethodOptions = append(
			paymentMethodOptions,
			sheet.NewSelectOption(
				localization.paymentMethodLabels[paymentMethod],
				sheet.WithColor(entity.PaymentMethodColors[paymentMethod]),
			),
		)
	}

	return []sheet.CreateTableOption{
		sheet.WithIcon(transactionTableIcon),
		sheet.WithColumns(
			sheet.NewTitleColumn(columns.name),
			sheet.NewSelectColumn(
				columns.category,
				sheet.WithSelectOptions(categoryOptions...),
			),
			sheet.NewNumberColumn(
				columns.amount,
				sheet.WithCurrency(transactionCurrency),
			),
			sheet.NewSelectColumn(
				columns.paymentMethod,
				sheet.WithSelectOptions(paymentMethodOptions...),
			),
			sheet.NewTextColumn(columns.cardLastDigits),
			sheet.NewDateColumn(columns.date),
		),
	}
}

func transactionToRow(transaction entity.Transaction, language entity.Language) sheet.Row {
	localization := transactionTableLocalizationFor(language)
	columns := localization.columns

	cardLastDigits := ""
	if transaction.CardLastDigits != nil {
		cardLastDigits = *transaction.CardLastDigits
	}

	return sheet.Row{
		columns.name:           sheet.TitleCell(transaction.Name),
		columns.category:       sheet.SelectCell(transaction.Category),
		columns.amount:         sheet.NumberCell(transaction.Amount),
		columns.paymentMethod:  sheet.SelectCell(localization.paymentMethodLabels[transaction.PaymentMethod]),
		columns.cardLastDigits: sheet.TextCell(cardLastDigits),
		columns.date:           sheet.DateCell(transaction.Date),
	}
}

func rowToTransaction(row sheet.Row, language entity.Language) (entity.Transaction, error) {
	localization := transactionTableLocalizationFor(language)
	columns := localization.columns

	name, err := rowCell[sheet.TitleCell](row, columns.name, sheet.ColumnTypeTitle)
	if err != nil {
		return entity.Transaction{}, err
	}

	category, err := rowCell[sheet.SelectCell](row, columns.category, sheet.ColumnTypeSelect)
	if err != nil {
		return entity.Transaction{}, err
	}

	amount, err := rowCell[sheet.NumberCell](row, columns.amount, sheet.ColumnTypeNumber)
	if err != nil {
		return entity.Transaction{}, err
	}

	paymentMethodLabel, err := rowCell[sheet.SelectCell](row, columns.paymentMethod, sheet.ColumnTypeSelect)
	if err != nil {
		return entity.Transaction{}, err
	}
	paymentMethod, exists := localization.paymentMethodsByLabel[string(paymentMethodLabel)]
	if !exists && paymentMethodLabel != "" {
		return entity.Transaction{}, fmt.Errorf(
			"column %q has unknown payment method %q",
			columns.paymentMethod,
			paymentMethodLabel,
		)
	}

	cardLastDigits, err := rowCell[sheet.TextCell](row, columns.cardLastDigits, sheet.ColumnTypeText)
	if err != nil {
		return entity.Transaction{}, err
	}

	date, err := rowCell[sheet.DateCell](row, columns.date, sheet.ColumnTypeDate)
	if err != nil {
		return entity.Transaction{}, err
	}

	transaction := entity.Transaction{
		Name:          string(name),
		Category:      entity.Category(category),
		Amount:        float64(amount),
		PaymentMethod: paymentMethod,
		Date:          time.Time(date),
	}
	if cardLastDigits != "" {
		value := string(cardLastDigits)
		transaction.CardLastDigits = &value
	}

	return transaction, nil
}

func rowCell[T sheet.Cell](row sheet.Row, column string, want sheet.ColumnType) (T, error) {
	var zero T
	cell, exists := row[column]
	if !exists {
		return zero, nil
	}

	value, ok := cell.(T)
	if !ok {
		return zero, fmt.Errorf("column %q has cell type %T, want %s", column, cell, want)
	}

	return value, nil
}
