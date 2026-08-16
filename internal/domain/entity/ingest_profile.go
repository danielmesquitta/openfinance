package entity

type IngestProfile struct {
	ID                        string                   `json:"id"                                     validate:"required"`
	Language                  Language                 `json:"language,omitempty"                     validate:"omitempty,oneof=en pt-BR"`
	NotionToken               string                   `json:"notion_token"                           validate:"required"`
	NotionPageID              string                   `json:"notion_page_id"                         validate:"required"`
	PluggyClientID            string                   `json:"pluggy_client_id"                       validate:"required"`
	PluggyClientSecret        string                   `json:"pluggy_client_secret"                   validate:"required"`
	PluggyAccountIDs          []string                 `json:"pluggy_account_ids"                     validate:"required,min=1,dive,required"`
	IgnoreSamePersonTransfers *bool                    `json:"ignore_same_person_transfers,omitempty"`
	Categories                map[Category]Color       `json:"categories"                             validate:"required,min=1,dive,keys,required,endkeys,required"`
	CategoryMappings          map[string]Category      `json:"category_mappings"                      validate:"required,dive,keys,required,endkeys,required"`
	Fallback                  Category                 `json:"fallback,omitempty"`
	BudgetGroups              map[BudgetGroup]Color    `json:"budget_groups,omitempty"                validate:"omitempty,min=1,dive,keys,required,endkeys,required"`
	BudgetGroupMappings       map[Category]BudgetGroup `json:"budget_group_mappings,omitempty"        validate:"omitempty,dive,keys,required,endkeys,required"`
	BudgetGroupFallback       BudgetGroup              `json:"budget_group_fallback,omitempty"`
}
