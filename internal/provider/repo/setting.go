package repo

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type SettingRepo interface {
	CreateSetting(setting *entity.Setting) error
	UpdateSetting(id string, setting *entity.Setting) error
}
