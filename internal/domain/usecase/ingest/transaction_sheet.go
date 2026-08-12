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

	transactionNameColumn           = "Name"
	transactionCategoryColumn       = "Category"
	transactionAmountColumn         = "Amount"
	transactionPaymentMethodColumn  = "Payment Method"
	transactionCardLastDigitsColumn = "Card Last Digits"
	transactionDateColumn           = "Date"
)

func transactionTableOptions(settings entity.IngestProfileSettings) []sheet.CreateTableOption {
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
				string(paymentMethod),
				sheet.WithColor(entity.PaymentMethodColors[paymentMethod]),
			),
		)
	}

	return []sheet.CreateTableOption{
		sheet.WithIcon(transactionTableIcon),
		sheet.WithColumns(
			sheet.NewTitleColumn(transactionNameColumn),
			sheet.NewSelectColumn(
				transactionCategoryColumn,
				sheet.WithSelectOptions(categoryOptions...),
			),
			sheet.NewNumberColumn(
				transactionAmountColumn,
				sheet.WithCurrency(transactionCurrency),
			),
			sheet.NewSelectColumn(
				transactionPaymentMethodColumn,
				sheet.WithSelectOptions(paymentMethodOptions...),
			),
			sheet.NewTextColumn(transactionCardLastDigitsColumn),
			sheet.NewDateColumn(transactionDateColumn),
		),
	}
}

func transactionToRow(transaction entity.Transaction) sheet.Row {
	cardLastDigits := ""
	if transaction.CardLastDigits != nil {
		cardLastDigits = *transaction.CardLastDigits
	}

	return sheet.Row{
		transactionNameColumn:           sheet.TitleCell(transaction.Name),
		transactionCategoryColumn:       sheet.SelectCell(transaction.Category),
		transactionAmountColumn:         sheet.NumberCell(transaction.Amount),
		transactionPaymentMethodColumn:  sheet.SelectCell(transaction.PaymentMethod),
		transactionCardLastDigitsColumn: sheet.TextCell(cardLastDigits),
		transactionDateColumn:           sheet.DateCell(transaction.Date),
	}
}

func rowToTransaction(row sheet.Row) (entity.Transaction, error) {
	name, err := rowCell[sheet.TitleCell](row, transactionNameColumn, sheet.ColumnTypeTitle)
	if err != nil {
		return entity.Transaction{}, err
	}

	category, err := rowCell[sheet.SelectCell](row, transactionCategoryColumn, sheet.ColumnTypeSelect)
	if err != nil {
		return entity.Transaction{}, err
	}

	amount, err := rowCell[sheet.NumberCell](row, transactionAmountColumn, sheet.ColumnTypeNumber)
	if err != nil {
		return entity.Transaction{}, err
	}

	paymentMethod, err := rowCell[sheet.SelectCell](row, transactionPaymentMethodColumn, sheet.ColumnTypeSelect)
	if err != nil {
		return entity.Transaction{}, err
	}

	cardLastDigits, err := rowCell[sheet.TextCell](row, transactionCardLastDigitsColumn, sheet.ColumnTypeText)
	if err != nil {
		return entity.Transaction{}, err
	}

	date, err := rowCell[sheet.DateCell](row, transactionDateColumn, sheet.ColumnTypeDate)
	if err != nil {
		return entity.Transaction{}, err
	}

	transaction := entity.Transaction{
		Name:          string(name),
		Category:      entity.Category(category),
		Amount:        float64(amount),
		PaymentMethod: entity.PaymentMethod(paymentMethod),
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
