package entity

type IngestProfile struct {
	ID                 string              `json:"id"`
	NotionToken        string              `json:"notion_token"`
	NotionPageID       string              `json:"notion_page_id"`
	PluggyClientID     string              `json:"pluggy_client_id"`
	PluggyClientSecret string              `json:"pluggy_client_secret"`
	PluggyAccountIDs   []string            `json:"pluggy_account_ids"`
	Categories         map[Category]Color  `json:"categories"`
	Mappings           map[string]Category `json:"mappings"`
	Fallback           Category            `json:"fallback,omitempty"`
}
