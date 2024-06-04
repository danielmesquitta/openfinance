package repo

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type CreateSettingDTO struct {
	NotionToken           string   `json:"notion_token"             validate:"required"`
	NotionPageID          string   `json:"notion_page_id"           validate:"required"`
	MeuPluggyClientID     string   `json:"meu_pluggy_client_id"     validate:"required"`
	MeuPluggyClientSecret string   `json:"meu_pluggy_client_secret" validate:"required"`
	MeuPluggyAccountIDs   []string `json:"meu_pluggy_account_ids"   validate:"required"`
	UserID                string   `json:"user_id"                  validate:"required"`
}

type UpdateSettingDTO struct {
	NotionToken           string   `json:"notion_token"             validate:"required"`
	NotionPageID          string   `json:"notion_page_id"           validate:"required"`
	MeuPluggyClientID     string   `json:"meu_pluggy_client_id"     validate:"required"`
	MeuPluggyClientSecret string   `json:"meu_pluggy_client_secret" validate:"required"`
	MeuPluggyAccountIDs   []string `json:"meu_pluggy_account_ids"   validate:"required"`
}

type SettingRepo interface {
	CreateSetting(dto CreateSettingDTO) (entity.Setting, error)
	UpdateSetting(id string, dto UpdateSettingDTO) (entity.Setting, error)
	ListSettings() ([]entity.Setting, error)
}
