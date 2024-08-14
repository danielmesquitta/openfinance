package envrepo

import (
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
)

type SettingEnvRepo struct {
	env    *Env
	cripto crypto.Encrypter
}

func NewSettingEnvRepo(
	env *Env,
	cripto crypto.Encrypter,
) *SettingEnvRepo {
	return &SettingEnvRepo{
		env:    env,
		cripto: cripto,
	}
}

func (s SettingEnvRepo) CreateSetting(
	_ repo.CreateSettingDTO,
) (entity.Setting, error) {
	panic("not implemented")
}

func (s SettingEnvRepo) UpdateSetting(
	_ string,
	_ repo.UpdateSettingDTO,
) (entity.Setting, error) {
	panic("not implemented")
}

func (s SettingEnvRepo) ListSettings() ([]entity.Setting, error) {
	return []entity.Setting{
		getDefaultSetting(s.env, s.cripto),
	}, nil
}

var _ repo.SettingRepo = (*SettingEnvRepo)(nil)
