package jsonrepo

import (
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
)

type SettingJSONRepo struct {
	cripto crypto.Encrypter
}

func (s SettingJSONRepo) CreateSetting(
	_ repo.CreateSettingDTO,
) (entity.Setting, error) {
	panic("not implemented") // TODO: Implement
}

func (s SettingJSONRepo) UpdateSetting(
	_ string,
	_ repo.UpdateSettingDTO,
) (entity.Setting, error) {
	panic("not implemented") // TODO: Implement
}

func (s SettingJSONRepo) ListSettings() ([]entity.Setting, error) {
	users := getDefaultUsers(s.cripto)
	settings := make([]entity.Setting, len(users))
	for i, u := range users {
		if u.Setting == nil {
			return nil, fmt.Errorf("Setting is nil for user %s", u.ID)
		}
		settings[i] = *u.Setting
	}
	return settings, nil
}

func NewSettingJSONRepo(
	cripto crypto.Encrypter,
) *SettingJSONRepo {
	return &SettingJSONRepo{
		cripto: cripto,
	}
}
