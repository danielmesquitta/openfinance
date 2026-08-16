package entity

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type IngestProfileSettings struct {
	ID                        string
	Language                  Language
	IgnoreSamePersonTransfers bool
	Categories                []Category
	ColorsByCategory          map[Category]Color
	Mappings                  map[string]Category
	Fallback                  Category
	BudgetGroups              []BudgetGroup
	ColorsByBudgetGroup       map[BudgetGroup]Color
	BudgetGroupMappings       map[Category]BudgetGroup
	BudgetGroupFallback       BudgetGroup
}

type IngestSettings struct {
	IngestProfiles []IngestProfileSettings
}

func NewIngestSettings(ingestProfiles []IngestProfile) (IngestSettings, error) {
	if err := ValidateIngestProfiles(ingestProfiles); err != nil {
		return IngestSettings{}, err
	}

	ingestProfileSettings := make([]IngestProfileSettings, 0, len(ingestProfiles))
	for _, ingestProfile := range ingestProfiles {
		settings, err := newIngestProfileSettings(ingestProfile)
		if err != nil {
			return IngestSettings{}, err
		}

		ingestProfileSettings = append(ingestProfileSettings, settings)
	}

	return IngestSettings{IngestProfiles: ingestProfileSettings}, nil
}

func newIngestProfileSettings(ingestProfile IngestProfile) (IngestProfileSettings, error) {
	language := ingestProfile.Language
	if language == "" {
		language = DefaultLanguage
	}
	if !language.IsValid() {
		return IngestProfileSettings{}, fmt.Errorf(
			"ingest profile %q: unsupported language %q (supported: %s, %s)",
			ingestProfile.ID,
			language,
			LanguageEnglish,
			LanguagePortugueseBrazil,
		)
	}

	ignoreSamePersonTransfers := true
	if ingestProfile.IgnoreSamePersonTransfers != nil {
		ignoreSamePersonTransfers = *ingestProfile.IgnoreSamePersonTransfers
	}

	if len(ingestProfile.Categories) == 0 {
		return IngestProfileSettings{}, fmt.Errorf(
			"ingest profile %q: at least one category is required",
			ingestProfile.ID,
		)
	}
	if ingestProfile.CategoryMappings == nil {
		return IngestProfileSettings{}, fmt.Errorf(
			"ingest profile %q: category mappings are required",
			ingestProfile.ID,
		)
	}

	fallback := ingestProfile.Fallback
	if fallback == "" {
		fallback = DefaultFallbackCategory
	}

	colorsByCategory := maps.Clone(ingestProfile.Categories)
	if _, exists := colorsByCategory[fallback]; !exists {
		colorsByCategory[fallback] = Gray
	}

	if err := ValidateCategories(colorsByCategory, ingestProfile.CategoryMappings); err != nil {
		return IngestProfileSettings{}, fmt.Errorf("ingest profile %q: %w", ingestProfile.ID, err)
	}

	categories := make([]Category, 0, len(colorsByCategory))
	for category := range colorsByCategory {
		categories = append(categories, category)
	}
	slices.Sort(categories)

	budgetGroupSettings, err := normalizeBudgetGroups(ingestProfile, colorsByCategory)
	if err != nil {
		return IngestProfileSettings{}, fmt.Errorf("ingest profile %q: %w", ingestProfile.ID, err)
	}

	return IngestProfileSettings{
		ID:                        ingestProfile.ID,
		Language:                  language,
		IgnoreSamePersonTransfers: ignoreSamePersonTransfers,
		Categories:                categories,
		ColorsByCategory:          colorsByCategory,
		Mappings:                  maps.Clone(ingestProfile.CategoryMappings),
		Fallback:                  fallback,
		BudgetGroups:              budgetGroupSettings.groups,
		ColorsByBudgetGroup:       budgetGroupSettings.colorsByGroup,
		BudgetGroupMappings:       budgetGroupSettings.mappings,
		BudgetGroupFallback:       budgetGroupSettings.fallback,
	}, nil
}

type normalizedBudgetGroupSettings struct {
	groups        []BudgetGroup
	colorsByGroup map[BudgetGroup]Color
	mappings      map[Category]BudgetGroup
	fallback      BudgetGroup
}

func normalizeBudgetGroups(
	ingestProfile IngestProfile,
	colorsByCategory map[Category]Color,
) (normalizedBudgetGroupSettings, error) {
	if ingestProfile.BudgetGroups == nil {
		if ingestProfile.BudgetGroupMappings != nil || ingestProfile.BudgetGroupFallback != "" {
			return normalizedBudgetGroupSettings{}, errors.New(
				"budget groups are required when budget group mappings or fallback are configured",
			)
		}

		return normalizedBudgetGroupSettings{}, nil
	}

	if len(ingestProfile.BudgetGroups) == 0 {
		return normalizedBudgetGroupSettings{}, errors.New("at least one budget group is required")
	}
	if ingestProfile.BudgetGroupMappings == nil {
		return normalizedBudgetGroupSettings{}, errors.New("budget group mappings are required")
	}

	fallback := ingestProfile.BudgetGroupFallback
	if fallback == "" {
		fallback = DefaultFallbackBudgetGroup
	}

	colorsByBudgetGroup := maps.Clone(ingestProfile.BudgetGroups)
	if _, exists := colorsByBudgetGroup[fallback]; !exists {
		colorsByBudgetGroup[fallback] = Gray
	}

	if err := ValidateBudgetGroups(
		colorsByCategory,
		colorsByBudgetGroup,
		ingestProfile.BudgetGroupMappings,
	); err != nil {
		return normalizedBudgetGroupSettings{}, err
	}

	budgetGroups := make([]BudgetGroup, 0, len(colorsByBudgetGroup))
	for budgetGroup := range colorsByBudgetGroup {
		budgetGroups = append(budgetGroups, budgetGroup)
	}
	slices.Sort(budgetGroups)

	return normalizedBudgetGroupSettings{
		groups:        budgetGroups,
		colorsByGroup: colorsByBudgetGroup,
		mappings:      maps.Clone(ingestProfile.BudgetGroupMappings),
		fallback:      fallback,
	}, nil
}

func ValidateIngestProfiles(ingestProfiles []IngestProfile) error {
	if len(ingestProfiles) == 0 {
		return errors.New("at least one ingest profile is required")
	}

	ingestProfileIDs := make(map[string]struct{}, len(ingestProfiles))
	for _, ingestProfile := range ingestProfiles {
		if ingestProfile.ID == "" || ingestProfile.NotionToken == "" || ingestProfile.NotionPageID == "" ||
			ingestProfile.PluggyClientID == "" || ingestProfile.PluggyClientSecret == "" ||
			len(ingestProfile.PluggyAccountIDs) == 0 {
			return fmt.Errorf("ingest profile %q has incomplete integration settings", ingestProfile.ID)
		}

		if _, exists := ingestProfileIDs[ingestProfile.ID]; exists {
			return fmt.Errorf("ingest profile id %q is duplicated", ingestProfile.ID)
		}

		ingestProfileIDs[ingestProfile.ID] = struct{}{}
	}

	return nil
}

func ValidateCategories(
	colorsByCategory map[Category]Color,
	mappings map[string]Category,
) error {
	if len(colorsByCategory) == 0 {
		return errors.New("at least one category is required")
	}

	for category, color := range colorsByCategory {
		if category == "" {
			return errors.New("category name cannot be empty")
		}

		if !color.IsValid() {
			return fmt.Errorf("color %q for category %q is invalid", color, category)
		}
	}

	for transactionName, category := range mappings {
		if transactionName == "" {
			return errors.New("mapping transaction name cannot be empty")
		}

		if _, ok := colorsByCategory[category]; !ok {
			return fmt.Errorf("mapping category %q is not configured", category)
		}
	}

	return nil
}

func ValidateBudgetGroups(
	colorsByCategory map[Category]Color,
	colorsByBudgetGroup map[BudgetGroup]Color,
	mappings map[Category]BudgetGroup,
) error {
	if len(colorsByBudgetGroup) == 0 {
		return errors.New("at least one budget group is required")
	}

	for budgetGroup, color := range colorsByBudgetGroup {
		if budgetGroup == "" {
			return errors.New("budget group name cannot be empty")
		}

		if !color.IsValid() {
			return fmt.Errorf("color %q for budget group %q is invalid", color, budgetGroup)
		}
	}

	for category, budgetGroup := range mappings {
		if category == "" {
			return errors.New("budget group mapping category cannot be empty")
		}

		if _, ok := colorsByCategory[category]; !ok {
			return fmt.Errorf("mapping category %q is not configured", category)
		}

		if _, ok := colorsByBudgetGroup[budgetGroup]; !ok {
			return fmt.Errorf("mapping budget group %q is not configured", budgetGroup)
		}
	}

	return nil
}
