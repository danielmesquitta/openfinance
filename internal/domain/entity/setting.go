package entity

import "time"

type Setting struct {
	ID                    string    `json:"id"`
	NotionToken           string    `json:"notion_token"`
	NotionPageID          string    `json:"notion_page_id"`
	MeuPluggyClientID     string    `json:"meu_pluggy_client_id"`
	MeuPluggyClientSecret string    `json:"meu_pluggy_client_secret"`
	MeuPluggyAccountIDs   []string  `json:"meu_pluggy_account_ids"`
	UserID                string    `json:"user_id"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (s *Setting) Encrypt() {
}

func (s *Setting) Decrypt() {
}
