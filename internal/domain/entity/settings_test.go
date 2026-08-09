package entity

import "testing"

func TestNewSyncSettings(t *testing.T) {
	t.Parallel()

	users := []User{{
		ID:                 "user",
		NotionToken:        "token",
		NotionPageID:       "page",
		PluggyClientID:     "client",
		PluggyClientSecret: "secret",
		PluggyAccountIDs:   []string{"account"},
	}}
	colors := map[Category]Color{CategoryUnknown: Gray, "Food": Red}
	mappings := map[string]Category{"Market": "Food"}

	settings, err := NewSyncSettings(users, colors, mappings)
	if err != nil {
		t.Fatalf("NewSyncSettings() error = %v", err)
	}

	if len(settings.UserIDs) != 1 || settings.UserIDs[0] != "user" {
		t.Fatalf("user IDs = %#v", settings.UserIDs)
	}
	if len(settings.Categories) != 2 || settings.Categories[0] != "Food" || settings.Categories[1] != CategoryUnknown {
		t.Fatalf("categories = %#v", settings.Categories)
	}
	if settings.Fallback != CategoryUnknown {
		t.Fatalf("fallback = %q", settings.Fallback)
	}
}

func TestSettingsValidation(t *testing.T) {
	t.Parallel()

	validUser := User{
		ID:                 "user",
		NotionToken:        "token",
		NotionPageID:       "page",
		PluggyClientID:     "client",
		PluggyClientSecret: "secret",
		PluggyAccountIDs:   []string{"account"},
	}

	tests := []struct {
		name     string
		users    []User
		colors   map[Category]Color
		mappings map[string]Category
	}{
		{name: "no users", colors: map[Category]Color{CategoryUnknown: Gray}},
		{
			name:   "duplicate user",
			users:  []User{validUser, validUser},
			colors: map[Category]Color{CategoryUnknown: Gray},
		},
		{name: "missing fallback", users: []User{validUser}, colors: map[Category]Color{"Food": Red}},
		{name: "invalid color", users: []User{validUser}, colors: map[Category]Color{CategoryUnknown: "invalid"}},
		{
			name:     "unknown mapping category",
			users:    []User{validUser},
			colors:   map[Category]Color{CategoryUnknown: Gray},
			mappings: map[string]Category{"Store": "Shopping"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewSyncSettings(test.users, test.colors, test.mappings); err == nil {
				t.Fatal("NewSyncSettings() error = nil")
			}
		})
	}
}
