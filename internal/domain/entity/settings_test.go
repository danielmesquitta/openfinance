package entity

import "testing"

func validIngestProfile() IngestProfile {
	return IngestProfile{
		ID:                 "ingest-profile",
		NotionToken:        "token",
		NotionPageID:       "page",
		PluggyClientID:     "client",
		PluggyClientSecret: "secret",
		PluggyAccountIDs:   []string{"account"},
		Categories:         map[Category]Color{"Food": Red},
		Mappings:           map[string]Category{},
	}
}

func TestNewIngestSettingsNormalizesEachIngestProfile(t *testing.T) {
	t.Parallel()

	first := validIngestProfile()
	first.Mappings = map[string]Category{"Market": "Food"}
	first.IgnoreSamePersonTransfers = new(false)
	second := validIngestProfile()
	second.ID = "second"
	second.Categories = map[Category]Color{"Education": Blue}
	second.Mappings = map[string]Category{"Unknown store": "Outros"}
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
	t.Parallel()

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

func TestNewIngestSettingsValidation(t *testing.T) {
	t.Parallel()

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
			name: "missing mappings",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Mappings = nil

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
				ingestProfile.Mappings[""] = "Food"

				return []IngestProfile{ingestProfile}
			},
		},
		{
			name: "unknown mapping category",
			ingestProfiles: func() []IngestProfile {
				ingestProfile := validIngestProfile()
				ingestProfile.Mappings["Store"] = "Shopping"

				return []IngestProfile{ingestProfile}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewIngestSettings(test.ingestProfiles()); err == nil {
				t.Fatal("NewIngestSettings() error = nil")
			}
		})
	}
}

func TestNewIngestSettingsUnsupportedLanguageError(t *testing.T) {
	t.Parallel()

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
