package pgrepo

import (
	"context"
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

func (b SettingPgRepo) CreateSetting(setting *entity.Setting) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := pgdb.CreateSettingParams{}

	err := copier.Copy(&params, setting)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	params.ID = id.String()
	params.UpdatedAt = time.Now()

	err = b.db.CreateSetting(ctx, params)
	if err != nil {
		return err
	}

	setting.ID = params.ID

	return nil
}

func (b SettingPgRepo) UpdateSetting(id string, setting *entity.Setting) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := pgdb.UpdateSettingParams{}
	err := copier.Copy(&params, setting)
	if err != nil {
		return err
	}

	params.ID = id
	params.UpdatedAt = time.Now()

	err = b.db.UpdateSetting(ctx, params)
	if err != nil {
		return err
	}

	setting.ID = params.ID

	return nil
}
