package usecase

import (
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/jinzhu/copier"
)

type UpsertUserSettingUseCase struct {
	userRepo    repo.UserRepo
	settingRepo repo.SettingRepo
	val         *validator.Validator
	cripto      crypto.Encrypter
}

func NewUpsertUserSettingUseCase(
	userRepo repo.UserRepo,
	settingRepo repo.SettingRepo,
	val *validator.Validator,
	cripto crypto.Encrypter,
) *UpsertUserSettingUseCase {
	return &UpsertUserSettingUseCase{
		userRepo:    userRepo,
		settingRepo: settingRepo,
		val:         val,
		cripto:      cripto,
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
	if err := uc.validateUserID(dto.UserID); err != nil {
		return err
	}

	user, err := uc.userRepo.GetUserWithSettingByID(dto.UserID)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return entity.ErrUserNotFound
	}

	if err := uc.encryptDTO(&dto); err != nil {
		return err
	}

	setting := user.Setting

	if settingNotExists := setting.ID == ""; settingNotExists {
		return uc.createSetting(setting, dto)
	}

	return uc.updateSetting(setting, dto)
}

func (uc *UpsertUserSettingUseCase) validateUserID(
	userID string,
) error {
	if userID == "" {
		err := entity.ErrValidation
		err.Message = "user_id is required"
		return &err
	}

	return nil
}

func (uc *UpsertUserSettingUseCase) updateSetting(
	setting *entity.Setting,
	dto UpsertUserSettingDTO,
) error {
	if err := copier.CopyWithOption(
		setting,
		dto,
		copier.Option{IgnoreEmpty: true},
	); err != nil {
		return fmt.Errorf("error copying dto to setting: %w", err)
	}

	if err := uc.val.Validate(setting); err != nil {
		return err
	}

	if err := uc.settingRepo.UpdateSetting(setting.ID, setting); err != nil {
		return fmt.Errorf("error updating setting: %w", err)
	}

	return nil
}

func (uc *UpsertUserSettingUseCase) createSetting(
	setting *entity.Setting,
	dto UpsertUserSettingDTO,
) error {
	if err := uc.val.Validate(dto); err != nil {
		return err
	}

	if err := copier.Copy(setting, dto); err != nil {
		return fmt.Errorf("error copying dto to setting: %w", err)
	}

	if err := uc.settingRepo.CreateSetting(setting); err != nil {
		return fmt.Errorf("error creating setting: %w", err)
	}

	return nil
}

func (uc *UpsertUserSettingUseCase) encryptDTO(
	dto *UpsertUserSettingDTO,
) error {
	for i, accountID := range dto.MeuPluggyAccountIDs {
		hashed, err := uc.cripto.Encrypt(accountID)
		if err != nil {
			return fmt.Errorf("error encrypting account_id: %w", err)
		}
		dto.MeuPluggyAccountIDs[i] = hashed
	}

	hashedMeuPluggyClientID, err := uc.cripto.Encrypt(dto.MeuPluggyClientID)
	if err != nil {
		return fmt.Errorf("error encrypting meu_pluggy_client_id: %w", err)
	}
	dto.MeuPluggyClientID = hashedMeuPluggyClientID

	hashedMeuPluggyClientSecret, err := uc.cripto.Encrypt(
		dto.MeuPluggyClientSecret,
	)
	if err != nil {
		return fmt.Errorf("error encrypting meu_pluggy_client_secret: %w", err)
	}
	dto.MeuPluggyClientSecret = hashedMeuPluggyClientSecret

	hashedNotionPageID, err := uc.cripto.Encrypt(dto.NotionPageID)
	if err != nil {
		return fmt.Errorf("error encrypting notion_page_id: %w", err)
	}
	dto.NotionPageID = hashedNotionPageID

	hashedNotionToken, err := uc.cripto.Encrypt(dto.NotionToken)
	if err != nil {
		return fmt.Errorf("error encrypting notion_token: %w", err)
	}
	dto.NotionToken = hashedNotionToken

	return nil
}
