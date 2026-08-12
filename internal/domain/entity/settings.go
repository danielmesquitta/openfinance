package entity

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type IngestProfileSettings struct {
	ID               string
	Language         Language
	Categories       []Category
	ColorsByCategory map[Category]Color
	Mappings         map[string]Category
	Fallback         Category
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
		language := ingestProfile.Language
		if language == "" {
			language = DefaultLanguage
		}
		if !language.IsValid() {
			return IngestSettings{}, fmt.Errorf(
				"ingest profile %q: unsupported language %q (supported: %s, %s)",
				ingestProfile.ID,
				language,
				LanguageEnglish,
				LanguagePortugueseBrazil,
			)
		}

		if len(ingestProfile.Categories) == 0 {
			return IngestSettings{}, fmt.Errorf("ingest profile %q: at least one category is required", ingestProfile.ID)
		}
		if ingestProfile.Mappings == nil {
			return IngestSettings{}, fmt.Errorf("ingest profile %q: mappings are required", ingestProfile.ID)
		}

		fallback := ingestProfile.Fallback
		if fallback == "" {
			fallback = DefaultFallbackCategory
		}

		colorsByCategory := maps.Clone(ingestProfile.Categories)
		if _, exists := colorsByCategory[fallback]; !exists {
			colorsByCategory[fallback] = Gray
		}

		if err := ValidateCategories(colorsByCategory, ingestProfile.Mappings); err != nil {
			return IngestSettings{}, fmt.Errorf("ingest profile %q: %w", ingestProfile.ID, err)
		}

		categories := make([]Category, 0, len(colorsByCategory))
		for category := range colorsByCategory {
			categories = append(categories, category)
		}
		slices.Sort(categories)

		ingestProfileSettings = append(ingestProfileSettings, IngestProfileSettings{
			ID:               ingestProfile.ID,
			Language:         language,
			Categories:       categories,
			ColorsByCategory: colorsByCategory,
			Mappings:         maps.Clone(ingestProfile.Mappings),
			Fallback:         fallback,
		})
	}

	return IngestSettings{IngestProfiles: ingestProfileSettings}, nil
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
