package entity

import "testing"

func validSyncProfile() SyncProfile {
	return SyncProfile{
		ID:                 "sync-profile",
		NotionToken:        "token",
		NotionPageID:       "page",
		PluggyClientID:     "client",
		PluggyClientSecret: "secret",
		PluggyAccountIDs:   []string{"account"},
		Categories:         map[Category]Color{"Food": Red},
		Mappings:           map[string]Category{},
	}
}

func TestNewSyncSettingsNormalizesEachSyncProfile(t *testing.T) {
	t.Parallel()

	first := validSyncProfile()
	first.Mappings = map[string]Category{"Market": "Food"}
	second := validSyncProfile()
	second.ID = "second"
	second.Categories = map[Category]Color{"Education": Blue}
	second.Mappings = map[string]Category{"Unknown store": "Outros"}
	second.Fallback = "Outros"

	settings, err := NewSyncSettings([]SyncProfile{first, second})
	if err != nil {
		t.Fatalf("NewSyncSettings() error = %v", err)
	}

	if len(settings.SyncProfiles) != 2 || settings.SyncProfiles[0].ID != "sync-profile" ||
		settings.SyncProfiles[1].ID != "second" {
		t.Fatalf("sync profiles = %#v", settings.SyncProfiles)
	}

	firstSettings := settings.SyncProfiles[0]
	if len(firstSettings.Categories) != 2 || firstSettings.Categories[0] != "Food" ||
		firstSettings.Categories[1] != DefaultFallbackCategory {
		t.Fatalf("first categories = %#v", firstSettings.Categories)
	}
	if firstSettings.Fallback != DefaultFallbackCategory ||
		firstSettings.ColorsByCategory[DefaultFallbackCategory] != Gray {
		t.Fatalf("first settings = %#v", firstSettings)
	}
	if _, mutated := first.Categories[DefaultFallbackCategory]; mutated {
		t.Fatal("NewSyncSettings() mutated the input categories")
	}

	secondSettings := settings.SyncProfiles[1]
	if len(secondSettings.Categories) != 2 || secondSettings.Categories[0] != "Education" ||
		secondSettings.Categories[1] != "Outros" {
		t.Fatalf("second categories = %#v", secondSettings.Categories)
	}
	if secondSettings.Fallback != "Outros" || secondSettings.ColorsByCategory["Outros"] != Gray {
		t.Fatalf("second settings = %#v", secondSettings)
	}
}

func TestNewSyncSettingsPreservesConfiguredFallbackColor(t *testing.T) {
	t.Parallel()

	syncProfile := validSyncProfile()
	syncProfile.Fallback = "Outros"
	syncProfile.Categories["Outros"] = Purple

	settings, err := NewSyncSettings([]SyncProfile{syncProfile})
	if err != nil {
		t.Fatalf("NewSyncSettings() error = %v", err)
	}
	if settings.SyncProfiles[0].ColorsByCategory["Outros"] != Purple {
		t.Fatalf("fallback color = %q", settings.SyncProfiles[0].ColorsByCategory["Outros"])
	}
}

func TestNewSyncSettingsValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		syncProfiles func() []SyncProfile
	}{
		{name: "no sync profiles", syncProfiles: func() []SyncProfile { return nil }},
		{
			name: "duplicate sync profile",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				return []SyncProfile{syncProfile, syncProfile}
			},
		},
		{
			name: "incomplete integration",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.NotionToken = ""
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "missing categories",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Categories = nil
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "missing mappings",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Mappings = nil
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "invalid color",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Categories["Food"] = "invalid"
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "empty category name",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Categories[""] = Red
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "empty mapping name",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Mappings[""] = "Food"
				return []SyncProfile{syncProfile}
			},
		},
		{
			name: "unknown mapping category",
			syncProfiles: func() []SyncProfile {
				syncProfile := validSyncProfile()
				syncProfile.Mappings["Store"] = "Shopping"
				return []SyncProfile{syncProfile}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewSyncSettings(test.syncProfiles()); err == nil {
				t.Fatal("NewSyncSettings() error = nil")
			}
		})
	}
}
