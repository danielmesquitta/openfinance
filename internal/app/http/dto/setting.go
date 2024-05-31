package dto

type UpsertUserSettingRequestDTO struct {
	NotionToken           string   `validate:"required"`
	NotionPageID          string   `validate:"required"`
	MeuPluggyClientID     string   `validate:"required"`
	MeuPluggyClientSecret string   `validate:"required"`
	MeuPluggyAccountIDs   []string `validate:"required"`
}
