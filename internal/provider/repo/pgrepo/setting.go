package pgrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/db/pgdb"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
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

func (s SettingPgRepo) CreateSetting(
	dto repo.CreateSettingDTO,
) (entity.Setting, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	arg := pgdb.CreateSettingParams{}
	if err := copier.Copy(&arg, dto); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying dto to db args: %w",
			err,
		)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return entity.Setting{}, fmt.Errorf("error generating uuid: %w", err)
	}

	arg.ID = id.String()
	arg.UpdatedAt = time.Now()

	result, err := s.db.CreateSetting(ctx, arg)
	if err != nil {
		return entity.Setting{}, fmt.Errorf("error creating setting: %w", err)
	}

	setting := entity.Setting{}
	if err := copier.Copy(&setting, result); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying result to setting entity: %w",
			err,
		)
	}

	return setting, nil
}

func (s SettingPgRepo) UpdateSetting(
	id string,
	dto repo.UpdateSettingDTO,
) (entity.Setting, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	arg := pgdb.UpdateSettingParams{}
	if err := copier.Copy(&arg, dto); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying dto to db args: %w",
			err,
		)
	}

	arg.ID = id
	arg.UpdatedAt = time.Now()

	result, err := s.db.UpdateSetting(ctx, arg)
	if err != nil {
		return entity.Setting{}, fmt.Errorf("error updating setting: %w", err)
	}

	setting := entity.Setting{}
	if err := copier.Copy(&setting, result); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying result to setting entity: %w",
			err,
		)
	}

	return setting, nil
}

func (s SettingPgRepo) ListSettings() ([]entity.Setting, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := s.db.ListSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing settings: %w", err)
	}

	var settings []entity.Setting
	if err := copier.Copy(&settings, result); err != nil {
		return nil, fmt.Errorf("error copying settings: %w", err)
	}

	return settings, nil
}
