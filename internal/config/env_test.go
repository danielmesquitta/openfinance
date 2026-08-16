package config

import (
	"strings"
	"testing"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/pkg/validator"
)

func validIngestProfilesFileData() IngestProfilesFileData {
	return IngestProfilesFileData{IngestProfiles: []entity.IngestProfile{{
		ID:                 "ingest-profile",
		NotionToken:        "notion-token",
		NotionPageID:       "notion-page",
		PluggyClientID:     "pluggy-client",
		PluggyClientSecret: "pluggy-secret",
		PluggyAccountIDs:   []string{"account"},
		Categories:         map[entity.Category]entity.Color{"Food": entity.Red},
		CategoryMappings:   map[string]entity.Category{},
	}}}
}

func TestIngestProfilesFileDataValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*IngestProfilesFileData)
		wantErr bool
	}{
		{name: "valid minimal configuration"},
		{
			name: "valid omitted language",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Language = ""
			},
		},
		{
			name: "valid empty category mappings",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].CategoryMappings = map[string]entity.Category{}
			},
		},
		{
			name: "valid omitted budget groups",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroups = nil
				data.IngestProfiles[0].BudgetGroupMappings = nil
			},
		},
		{
			name: "valid configured budget groups",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroups = map[entity.BudgetGroup]entity.Color{
					"Needs": entity.Red,
				}
				data.IngestProfiles[0].BudgetGroupMappings = map[entity.Category]entity.BudgetGroup{
					"Food": "Needs",
				}
			},
		},
		{
			name: "nil profiles",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles = nil
			},
			wantErr: true,
		},
		{
			name: "empty profiles",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles = []entity.IngestProfile{}
			},
			wantErr: true,
		},
		{
			name: "duplicate profile id",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles = append(data.IngestProfiles, data.IngestProfiles[0])
			},
			wantErr: true,
		},
		{
			name: "missing profile id",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].ID = ""
			},
			wantErr: true,
		},
		{
			name: "unsupported language",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Language = "es"
			},
			wantErr: true,
		},
		{
			name: "missing notion token",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].NotionToken = ""
			},
			wantErr: true,
		},
		{
			name: "missing notion page id",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].NotionPageID = ""
			},
			wantErr: true,
		},
		{
			name: "missing pluggy client id",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].PluggyClientID = ""
			},
			wantErr: true,
		},
		{
			name: "missing pluggy client secret",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].PluggyClientSecret = ""
			},
			wantErr: true,
		},
		{
			name: "nil pluggy accounts",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].PluggyAccountIDs = nil
			},
			wantErr: true,
		},
		{
			name: "empty pluggy accounts",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].PluggyAccountIDs = []string{}
			},
			wantErr: true,
		},
		{
			name: "blank pluggy account",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].PluggyAccountIDs = []string{""}
			},
			wantErr: true,
		},
		{
			name: "nil categories",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Categories = nil
			},
			wantErr: true,
		},
		{
			name: "empty categories",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Categories = map[entity.Category]entity.Color{}
			},
			wantErr: true,
		},
		{
			name: "empty category name",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Categories = map[entity.Category]entity.Color{"": entity.Red}
			},
			wantErr: true,
		},
		{
			name: "empty category color",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].Categories = map[entity.Category]entity.Color{"Food": ""}
			},
			wantErr: true,
		},
		{
			name: "nil category mappings",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].CategoryMappings = nil
			},
			wantErr: true,
		},
		{
			name: "empty category mapping name",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].CategoryMappings = map[string]entity.Category{"": "Food"}
			},
			wantErr: true,
		},
		{
			name: "empty mapped category",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].CategoryMappings = map[string]entity.Category{"Store": ""}
			},
			wantErr: true,
		},
		{
			name: "empty budget groups",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroups = map[entity.BudgetGroup]entity.Color{}
			},
			wantErr: true,
		},
		{
			name: "empty budget group name",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroups = map[entity.BudgetGroup]entity.Color{"": entity.Red}
			},
			wantErr: true,
		},
		{
			name: "empty budget group color",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroups = map[entity.BudgetGroup]entity.Color{"Needs": ""}
			},
			wantErr: true,
		},
		{
			name: "empty budget group mapping category",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroupMappings = map[entity.Category]entity.BudgetGroup{"": "Needs"}
			},
			wantErr: true,
		},
		{
			name: "empty mapped budget group",
			mutate: func(data *IngestProfilesFileData) {
				data.IngestProfiles[0].BudgetGroupMappings = map[entity.Category]entity.BudgetGroup{"Food": ""}
			},
			wantErr: true,
		},
	}

	val := validator.NewValidator()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := validIngestProfilesFileData()
			if test.mutate != nil {
				test.mutate(&data)
			}

			err := val.Validate(data)
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate(%#v) error = %v, wantErr %t", data, err, test.wantErr)
			}
		})
	}
}

func TestLoadIngestSettingsWrapsDomainValidationError(t *testing.T) {
	data := validIngestProfilesFileData()
	data.IngestProfiles[0].Categories["Food"] = "unsupported"
	e := Env{IngestProfilesFileData: data}

	err := e.loadIngestSettings()
	if err == nil || !strings.Contains(err.Error(), "invalid domain settings") {
		t.Fatalf("loadIngestSettings() error = %v, want invalid domain settings", err)
	}
}
