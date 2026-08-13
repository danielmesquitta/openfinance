package entity

type IngestProfile struct {
	ID                        string                 `json:"id"`
	Language                  Language               `json:"language,omitempty"`
	NotionToken               string                 `json:"notion_token"`
	NotionPageID              string                 `json:"notion_page_id"`
	PluggyClientID            string                 `json:"pluggy_client_id"`
	PluggyClientSecret        string                 `json:"pluggy_client_secret"`
	PluggyAccountIDs          []string               `json:"pluggy_account_ids"`
	IgnoreSamePersonTransfers *bool                  `json:"ignore_same_person_transfers,omitempty"`
	Categories                map[Category]Color     `json:"categories"`
	CategoryMappings          map[string]Category    `json:"category_mappings"`
	Fallback                  Category               `json:"fallback,omitempty"`
	BudgetGroups              map[BudgetGroup]Color  `json:"budget_groups,omitempty"`
	BudgetGroupMappings       map[string]BudgetGroup `json:"budget_group_mappings,omitempty"`
	BudgetGroupFallback       BudgetGroup            `json:"budget_group_fallback,omitempty"`
}
