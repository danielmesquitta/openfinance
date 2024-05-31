package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/db/pgdb"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type SettingPgRepo struct {
	db *pgdb.Queries
}

func NewSettingPgRepo(db *pgdb.Queries) *SettingPgRepo {
	return &SettingPgRepo{
		db: db,
	}
}

func (s SettingPgRepo) CreateSetting(setting *entity.Setting) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := pgdb.CreateSettingParams{}

	err := copier.Copy(&params, setting)
	if err != nil {
		return fmt.Errorf("error copying setting to params: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("error generating uuid: %w", err)
	}

	params.ID = id.String()
	params.UpdatedAt = time.Now()

	err = s.db.CreateSetting(ctx, params)
	if err != nil {
		return fmt.Errorf("error creating setting: %w", err)
	}

	setting.ID = params.ID

	return nil
}

func (s SettingPgRepo) UpdateSetting(id string, setting *entity.Setting) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := pgdb.UpdateSettingParams{}
	err := copier.Copy(&params, setting)
	if err != nil {
		return fmt.Errorf("error copying setting to params: %w", err)
	}

	params.ID = id
	params.UpdatedAt = time.Now()

	err = s.db.UpdateSetting(ctx, params)
	if err != nil {
		return fmt.Errorf("error updating setting: %w", err)
	}

	setting.ID = params.ID

	return nil
}

func (s SettingPgRepo) ListSettings() ([]entity.Setting, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dbSettings, err := s.db.ListSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing settings: %w", err)
	}

	var settings []entity.Setting
	err = copier.Copy(&settings, dbSettings)
	if err != nil {
		return nil, fmt.Errorf("error copying settings: %w", err)
	}

	return settings, nil
}
