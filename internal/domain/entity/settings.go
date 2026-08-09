package entity

import (
	"errors"
	"fmt"
	"slices"
)

type SyncSettings struct {
	UserIDs    []string
	Categories []Category
	Mappings   map[string]Category
	Fallback   Category
}

func NewSyncSettings(
	users []User,
	colorsByCategory map[Category]Color,
	mappings map[string]Category,
) (SyncSettings, error) {
	if err := ValidateUsers(users); err != nil {
		return SyncSettings{}, err
	}

	if err := ValidateCategories(colorsByCategory, mappings); err != nil {
		return SyncSettings{}, err
	}

	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	categories := make([]Category, 0, len(colorsByCategory))
	for category := range colorsByCategory {
		categories = append(categories, category)
	}
	slices.Sort(categories)

	return SyncSettings{
		UserIDs:    userIDs,
		Categories: categories,
		Mappings:   mappings,
		Fallback:   CategoryUnknown,
	}, nil
}

func ValidateUsers(users []User) error {
	if len(users) == 0 {
		return errors.New("at least one user is required")
	}

	userIDs := make(map[string]struct{}, len(users))
	for _, user := range users {
		if user.ID == "" || user.NotionToken == "" || user.NotionPageID == "" ||
			user.PluggyClientID == "" || user.PluggyClientSecret == "" || len(user.PluggyAccountIDs) == 0 {
			return fmt.Errorf("user %q has incomplete integration settings", user.ID)
		}

		if _, exists := userIDs[user.ID]; exists {
			return fmt.Errorf("user id %q is duplicated", user.ID)
		}

		userIDs[user.ID] = struct{}{}
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

	if _, ok := colorsByCategory[CategoryUnknown]; !ok {
		return fmt.Errorf("fallback category %q is required", CategoryUnknown)
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
