package dto

type UpsertUserSettingRequestDTO struct {
	NotionToken           string   `validate:"required" json:"notion_token,omitempty"`
	NotionPageID          string   `validate:"required" json:"notion_page_id,omitempty"`
	MeuPluggyClientID     string   `validate:"required" json:"meu_pluggy_client_id,omitempty"`
	MeuPluggyClientSecret string   `validate:"required" json:"meu_pluggy_client_secret,omitempty"`
	MeuPluggyAccountIDs   []string `validate:"required" json:"meu_pluggy_account_ids,omitempty"`
}
