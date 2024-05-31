package usecase

import (
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/hasher"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/jinzhu/copier"
)

type UpsertUserSettingUseCase struct {
	ur repo.UserRepo
	sr repo.SettingRepo
	v  *validator.Validator
	h  *hasher.Hasher
}

func NewUpsertUserSettingUseCase(
	ur repo.UserRepo,
	sr repo.SettingRepo,
	v *validator.Validator,
	h *hasher.Hasher,
) *UpsertUserSettingUseCase {
	return &UpsertUserSettingUseCase{
		ur: ur,
		sr: sr,
		v:  v,
		h:  h,
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
		err := entity.ErrValidation
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

	if err := uc.hashDTOValues(&dto); err != nil {
		return err
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

func (uc *UpsertUserSettingUseCase) hashDTOValues(
	dto *UpsertUserSettingDTO,
) error {
	for i, accountID := range dto.MeuPluggyAccountIDs {
		hashed, err := uc.h.Hash(accountID)
		if err != nil {
			return err
		}
		dto.MeuPluggyAccountIDs[i] = hashed
	}

	hashedMeuPluggyClientID, err := uc.h.Hash(dto.MeuPluggyClientID)
	if err != nil {
		return err
	}
	dto.MeuPluggyClientID = hashedMeuPluggyClientID

	hashedMeuPluggyClientSecret, err := uc.h.Hash(dto.MeuPluggyClientSecret)
	if err != nil {
		return err
	}
	dto.MeuPluggyClientSecret = hashedMeuPluggyClientSecret

	hashedNotionPageID, err := uc.h.Hash(dto.NotionPageID)
	if err != nil {
		return err
	}
	dto.NotionPageID = hashedNotionPageID

	hashedNotionToken, err := uc.h.Hash(dto.NotionToken)
	if err != nil {
		return err
	}
	dto.NotionToken = hashedNotionToken

	return nil
}
