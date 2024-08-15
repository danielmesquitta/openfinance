package entity

type User struct {
	ID                    string   `json:"id"`
	NotionToken           string   `json:"notion_token"`
	NotionPageID          string   `json:"notion_page_id"`
	MeuPluggyClientID     string   `json:"meu_pluggy_client_id"`
	MeuPluggyClientSecret string   `json:"meu_pluggy_client_secret"`
	MeuPluggyAccountIDs   []string `json:"meu_pluggy_account_ids"`
}
