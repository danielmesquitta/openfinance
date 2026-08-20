package ingest

import (
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
)

const (
	transactionTableIcon = "💸"
	transactionCurrency  = sheet.Currency("BRL")
)

func transactionTableDefinition(
	title string,
	settings entity.IngestProfileSettings,
) sheet.TableDefinition {
	localization := transactionTableLocalizationFor(settings.Language)
	columns := localization.columns

	categoryOptions := make([]sheet.SelectOption, 0, len(settings.Categories))
	for _, category := range settings.Categories {
		categoryOptions = append(
			categoryOptions,
			sheet.NewSelectOption(string(category)).
				Color(settings.ColorsByCategory[category]),
		)
	}

	paymentMethodOptions := make([]sheet.SelectOption, 0, len(entity.PaymentMethods))
	for _, paymentMethod := range entity.PaymentMethods {
		paymentMethodOptions = append(
			paymentMethodOptions,
			sheet.NewSelectOption(localization.paymentMethodLabels[paymentMethod]).
				Color(entity.PaymentMethodColors[paymentMethod]),
		)
	}

	definition := sheet.NewTable(title).
		SetIcon(transactionTableIcon).
		AddColumn(sheet.NewTitleColumn(columns.name)).
		AddColumn(sheet.NewSelectColumn(columns.category).Options(categoryOptions...))
	if budgetGroupColumn, enabled := budgetGroupTableColumn(settings, settings.Language); enabled {
		definition = definition.AddColumn(budgetGroupColumn)
	}

	return definition.
		AddColumn(
			sheet.NewNumberColumn(columns.amount).
				Currency(transactionCurrency),
		).
		AddColumn(
			sheet.NewSelectColumn(columns.paymentMethod).
				Options(paymentMethodOptions...),
		).
		AddColumn(sheet.NewTextColumn(columns.cardLastDigits)).
		AddColumn(sheet.NewDateColumn(columns.date))
}

func budgetGroupTableColumn(
	settings entity.IngestProfileSettings,
	language entity.Language,
) (sheet.SelectColumn, bool) {
	if len(settings.BudgetGroups) == 0 {
		return sheet.SelectColumn{}, false
	}

	options := make([]sheet.SelectOption, 0, len(settings.BudgetGroups))
	for _, budgetGroup := range settings.BudgetGroups {
		options = append(
			options,
			sheet.NewSelectOption(string(budgetGroup)).
				Color(settings.ColorsByBudgetGroup[budgetGroup]),
		)
	}

	columns := transactionTableLocalizationFor(language).columns

	return sheet.NewSelectColumn(columns.budgetGroup).Options(options...), true
}

func transactionToRow(transaction entity.Transaction, language entity.Language) sheet.Row {
	localization := transactionTableLocalizationFor(language)
	columns := localization.columns

	cardLastDigits := ""
	if transaction.CardLastDigits != nil {
		cardLastDigits = *transaction.CardLastDigits
	}

	row := sheet.Row{
		columns.name:           sheet.TitleCell(transaction.Name),
		columns.category:       sheet.SelectCell(transaction.Category),
		columns.amount:         sheet.NumberCell(transaction.Amount),
		columns.cardLastDigits: sheet.TextCell(cardLastDigits),
		columns.date:           sheet.DateCell(transaction.Date),
	}
	if transaction.BudgetGroup != "" {
		row[columns.budgetGroup] = sheet.SelectCell(transaction.BudgetGroup)
	}

	if paymentMethodLabel := localization.paymentMethodLabels[transaction.PaymentMethod]; paymentMethodLabel != "" {
		row[columns.paymentMethod] = sheet.SelectCell(paymentMethodLabel)
	}

	return row
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

	budgetGroup, err := rowCell[sheet.SelectCell](row, columns.budgetGroup, sheet.ColumnTypeSelect)
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
		BudgetGroup:   entity.BudgetGroup(budgetGroup),
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
