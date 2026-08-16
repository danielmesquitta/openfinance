package entity

import (
	"slices"
	"testing"
)

func validIngestProfile() IngestProfile {
	return IngestProfile{
		ID:                 "ingest-profile",
		NotionToken:        "token",
		NotionPageID:       "page",
		PluggyClientID:     "client",
		PluggyClientSecret: "secret",
		PluggyAccountIDs:   []string{"account"},
		Categories:         map[Category]Color{"Food": Red},
		CategoryMappings:   map[string]Category{},
	}
}

func TestNewIngestSettingsNormalizesEachIngestProfile(t *testing.T) {
	first := validIngestProfile()
	first.CategoryMappings = map[string]Category{"Market": "Food"}
	first.IgnoreSamePersonTransfers = new(false)
	second := validIngestProfile()
	second.ID = "second"
	second.Categories = map[Category]Color{"Education": Blue}
	second.CategoryMappings = map[string]Category{"Unknown store": "Outros"}
	second.Fallback = "Outros"
	second.Language = LanguagePortugueseBrazil

	settings, err := NewIngestSettings([]IngestProfile{first, second})
	if err != nil {
		t.Fatalf("NewIngestSettings() error = %v", err)
	}

	if len(settings.IngestProfiles) != 2 || settings.IngestProfiles[0].ID != "ingest-profile" ||
		settings.IngestProfiles[1].ID != "second" {
		t.Fatalf("ingest profiles = %#v", settings.IngestProfiles)
	}

	firstSettings := settings.IngestProfiles[0]
	if firstSettings.Language != LanguageEnglish {
		t.Fatalf("first language = %q, want %q", firstSettings.Language, LanguageEnglish)
	}
	if firstSettings.IgnoreSamePersonTransfers {
		t.Fatal("first profile ignores same-person transfers, want false")
	}
	if len(firstSettings.Categories) != 2 || firstSettings.Categories[0] != "Food" ||
		firstSettings.Categories[1] != DefaultFallbackCategory {
		t.Fatalf("first categories = %#v", firstSettings.Categories)
	}
	if firstSettings.Fallback != DefaultFallbackCategory ||
		firstSettings.ColorsByCategory[DefaultFallbackCategory] != Gray {
		t.Fatalf("first settings = %#v", firstSettings)
	}
	if _, mutated := first.Categories[DefaultFallbackCategory]; mutated {
		t.Fatal("NewIngestSettings() mutated the input categories")
	}

	secondSettings := settings.IngestProfiles[1]
	if secondSettings.Language != LanguagePortugueseBrazil {
		t.Fatalf(
			"second language = %q, want %q",
			secondSettings.Language,
			LanguagePortugueseBrazil,
		)
	}
	if !secondSettings.IgnoreSamePersonTransfers {
		t.Fatal("second profile ignores same-person transfers = false, want default true")
	}
	if len(secondSettings.Categories) != 2 || secondSettings.Categories[0] != "Education" ||
		secondSettings.Categories[1] != "Outros" {
		t.Fatalf("second categories = %#v", secondSettings.Categories)
	}
	if secondSettings.Fallback != "Outros" || secondSettings.ColorsByCategory["Outros"] != Gray {
		t.Fatalf("second settings = %#v", secondSettings)
	}
}

func TestNewIngestSettingsPreservesConfiguredFallbackColor(t *testing.T) {
	ingestProfile := validIngestProfile()
	ingestProfile.Fallback = "Outros"
	ingestProfile.Categories["Outros"] = Purple

	settings, err := NewIngestSettings([]IngestProfile{ingestProfile})
	if err != nil {
		t.Fatalf("NewIngestSettings() error = %v", err)
	}
	if settings.IngestProfiles[0].ColorsByCategory["Outros"] != Purple {
		t.Fatalf("fallback color = %q", settings.IngestProfiles[0].ColorsByCategory["Outros"])
	}
}

func TestNewIngestSettingsNormalizesBudgetGroups(t *testing.T) {
	ingestProfile := validIngestProfile()
	ingestProfile.BudgetGroups = map[BudgetGroup]Color{
		"Lifestyle":   Pink,
		"Fixed Costs": Red,
	}
	ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{"Food": "Fixed Costs"}

	settings, err := NewIngestSettings([]IngestProfile{ingestProfile})
	if err != nil {
		t.Fatalf("NewIngestSettings() error = %v", err)
	}

	got := settings.IngestProfiles[0]
	wantGroups := []BudgetGroup{"Fixed Costs", "Lifestyle", DefaultFallbackBudgetGroup}
	if !slices.Equal(got.BudgetGroups, wantGroups) {
		t.Fatalf("budget groups = %#v, want %#v", got.BudgetGroups, wantGroups)
	}
	if got.BudgetGroupFallback != DefaultFallbackBudgetGroup ||
		got.ColorsByBudgetGroup[DefaultFallbackBudgetGroup] != Gray ||
		got.BudgetGroupMappings["Food"] != "Fixed Costs" {
		t.Fatalf("budget group settings = %#v", got)
	}
	if _, mutated := ingestProfile.BudgetGroups[DefaultFallbackBudgetGroup]; mutated {
		t.Fatal("NewIngestSettings() mutated the input budget groups")
	}

	ingestProfile.BudgetGroups["Lifestyle"] = Blue
	ingestProfile.BudgetGroupMappings["Food"] = "Lifestyle"
	if got.ColorsByBudgetGroup["Lifestyle"] != Pink || got.BudgetGroupMappings["Food"] != "Fixed Costs" {
		t.Fatal("NewIngestSettings() retained mutable budget group input maps")
	}
}

func TestNewIngestSettingsPreservesConfiguredBudgetGroupFallbackColor(t *testing.T) {
	ingestProfile := validIngestProfile()
	ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": Red, "Unallocated": Purple}
	ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{}
	ingestProfile.BudgetGroupFallback = "Unallocated"

	settings, err := NewIngestSettings([]IngestProfile{ingestProfile})
	if err != nil {
		t.Fatalf("NewIngestSettings() error = %v", err)
	}
	got := settings.IngestProfiles[0]
	if got.BudgetGroupFallback != "Unallocated" || got.ColorsByBudgetGroup["Unallocated"] != Purple {
		t.Fatalf("budget group fallback settings = %#v", got)
	}
}

func TestNewIngestSettingsValidation(t *testing.T) {
	tests := []struct {
		name           string
		ingestProfiles func() []IngestProfile
	}{
		{name: "no ingest profiles", ingestProfiles: func() []IngestProfile { return nil }},
		{
			name: "duplicate ingest profile",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()

				return []IngestProfile{ingestProfile, ingestProfile}
			},
		},
		{
			name: "incomplete integration",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.NotionToken = ""

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "unsupported language",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Language = "es"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "missing categories",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Categories = nil

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "missing category mappings",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.CategoryMappings = nil

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "invalid color",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Categories["Food"] = "invalid"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "empty category name",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Categories[""] = Red

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "empty mapping name",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.CategoryMappings[""] = "Food"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "unknown mapping category",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.CategoryMappings["Store"] = "Shopping"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "empty budget groups",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "missing budget group mappings",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": Red}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "budget group mappings without groups",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "budget group fallback without groups",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroupFallback = "Other"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "invalid budget group color",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": "invalid"}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "empty budget group name",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"": Red}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "empty budget group mapping category",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": Red}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{"": "Needs"}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "unknown budget group mapping category",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": Red}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{"Shopping": "Needs"}

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "unknown mapped budget group",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.BudgetGroups = map[BudgetGroup]Color{"Needs": Red}
				ingestProfile.BudgetGroupMappings = map[Category]BudgetGroup{"Food": "Wants"}

				return []IngestProfile{ingestProfile}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewIngestSettings(test.ingestProfiles()); err == nil {
				t.Fatal("NewIngestSettings() error = nil")
			}
		})
	}
}

func TestNewIngestSettingsUnsupportedLanguageError(t *testing.T) {
	ingestProfile := validIngestProfile()
	ingestProfile.Language = "es"

	_, err := NewIngestSettings([]IngestProfile{ingestProfile})
	if err == nil {
		t.Fatal("NewIngestSettings() error = nil")
	}
	want := `ingest profile "ingest-profile": unsupported language "es" (supported: en, pt-BR)`
	if err.Error() != want {
		t.Fatalf("NewIngestSettings() error = %q, want %q", err, want)
	}
}
