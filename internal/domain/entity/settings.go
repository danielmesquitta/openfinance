package entity

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type SyncProfileSettings struct {
	ID               string
	Categories       []Category
	ColorsByCategory map[Category]Color
	Mappings         map[string]Category
	Fallback         Category
}

type SyncSettings struct {
	SyncProfiles []SyncProfileSettings
}

func NewSyncSettings(syncProfiles []SyncProfile) (SyncSettings, error) {
	if err := ValidateSyncProfiles(syncProfiles); err != nil {
		return SyncSettings{}, err
	}

	syncProfileSettings := make([]SyncProfileSettings, 0, len(syncProfiles))
	for _, syncProfile := range syncProfiles {
		if len(syncProfile.Categories) == 0 {
			return SyncSettings{}, fmt.Errorf("sync profile %q: at least one category is required", syncProfile.ID)
		}
		if syncProfile.Mappings == nil {
			return SyncSettings{}, fmt.Errorf("sync profile %q: mappings are required", syncProfile.ID)
		}

		fallback := syncProfile.Fallback
		if fallback == "" {
			fallback = DefaultFallbackCategory
		}

		colorsByCategory := maps.Clone(syncProfile.Categories)
		if _, exists := colorsByCategory[fallback]; !exists {
			colorsByCategory[fallback] = Gray
		}

		if err := ValidateCategories(colorsByCategory, syncProfile.Mappings); err != nil {
			return SyncSettings{}, fmt.Errorf("sync profile %q: %w", syncProfile.ID, err)
		}

		categories := make([]Category, 0, len(colorsByCategory))
		for category := range colorsByCategory {
			categories = append(categories, category)
		}
		slices.Sort(categories)

		syncProfileSettings = append(syncProfileSettings, SyncProfileSettings{
			ID:               syncProfile.ID,
			Categories:       categories,
			ColorsByCategory: colorsByCategory,
			Mappings:         maps.Clone(syncProfile.Mappings),
			Fallback:         fallback,
		})
	}

	return SyncSettings{SyncProfiles: syncProfileSettings}, nil
}

func ValidateSyncProfiles(syncProfiles []SyncProfile) error {
	if len(syncProfiles) == 0 {
		return errors.New("at least one sync profile is required")
	}

	syncProfileIDs := make(map[string]struct{}, len(syncProfiles))
	for _, syncProfile := range syncProfiles {
		if syncProfile.ID == "" || syncProfile.NotionToken == "" || syncProfile.NotionPageID == "" ||
			syncProfile.PluggyClientID == "" || syncProfile.PluggyClientSecret == "" ||
			len(syncProfile.PluggyAccountIDs) == 0 {
			return fmt.Errorf("sync profile %q has incomplete integration settings", syncProfile.ID)
		}

		if _, exists := syncProfileIDs[syncProfile.ID]; exists {
			return fmt.Errorf("sync profile id %q is duplicated", syncProfile.ID)
		}

		syncProfileIDs[syncProfile.ID] = struct{}{}
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
