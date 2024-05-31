package usecase

import (
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/jinzhu/copier"
)

type UpsertUserSettingUseCase struct {
	ur repo.UserRepo
	sr repo.SettingRepo
	v  *validator.Validator
}

func NewUpsertUserSettingUseCase(
	ur repo.UserRepo,
	sr repo.SettingRepo,
	v *validator.Validator,
) *UpsertUserSettingUseCase {
	return &UpsertUserSettingUseCase{
		ur: ur,
		sr: sr,
		v:  v,
	}
}

type UpsertUserSettingDTO struct {
	UserID                string   `validate:"required,uuid"`
	NotionToken           string   `validate:"required"`
	NotionPageID          string   `validate:"required"`
	MeuPluggyClientID     string   `validate:"required"`
	MeuPluggyClientSecret string   `validate:"required"`
	MeuPluggyAccountIDs   []string `validate:"required"`
}

func (uc *UpsertUserSettingUseCase) Execute(
	dto UpsertUserSettingDTO,
) error {
	if dto.UserID == "" {
		err := *entity.ErrValidation
		err.Message = "user_id is required"
		return &err
	}

	user, err := uc.ur.GetUserWithSettingByID(dto.UserID)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return entity.ErrUserNotFound
	}

	setting := user.Setting

	if settingNotExists := setting.ID == ""; settingNotExists {
		if err := uc.v.Validate(dto); err != nil {
			return err
		}

		if err := copier.Copy(setting, dto); err != nil {
			return err
		}

		if err := uc.sr.CreateSetting(setting); err != nil {
			return err
		}

		return nil
	}

	if err := copier.CopyWithOption(
		setting,
		dto,
		copier.Option{IgnoreEmpty: true},
	); err != nil {
		return err
	}

	if err := uc.v.Validate(setting); err != nil {
		return err
	}

	if err := uc.sr.UpdateSetting(setting.ID, setting); err != nil {
		return err
	}

	return nil
}
