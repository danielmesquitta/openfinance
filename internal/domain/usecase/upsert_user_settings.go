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
	UserID                string   `json:"user_id,omitempty"                  validate:"required,uuid"`
	NotionToken           string   `json:"notion_token,omitempty"`
	NotionPageID          string   `json:"notion_page_id,omitempty"`
	MeuPluggyClientID     string   `json:"meu_pluggy_client_id,omitempty"`
	MeuPluggyClientSecret string   `json:"meu_pluggy_client_secret,omitempty"`
	MeuPluggyAccountIDs   []string `json:"meu_pluggy_account_ids,omitempty"`
}

func (uc *UpsertUserSettingUseCase) Execute(
	dto UpsertUserSettingDTO,
) (entity.Setting, error) {
	if err := uc.val.Validate(dto); err != nil {
		return entity.Setting{}, err
	}

	user, err := uc.userRepo.GetFullUserByID(dto.UserID)
	if err != nil {
		return entity.Setting{}, err
	}
	if user.ID == "" {
		return entity.Setting{}, entity.ErrUserNotFound
	}

	if err := uc.encryptDTO(&dto); err != nil {
		return entity.Setting{}, err
	}

	setting := user.Setting

	if settingNotExists := setting.ID == ""; settingNotExists {
		return uc.createSetting(dto)
	}

	return uc.updateSetting(*setting, dto)
}

func (uc *UpsertUserSettingUseCase) updateSetting(
	setting entity.Setting,
	dto UpsertUserSettingDTO,
) (entity.Setting, error) {
	if err := copier.CopyWithOption(
		&setting,
		dto,
		copier.Option{IgnoreEmpty: true},
	); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying dto to setting: %w",
			err,
		)
	}

	params := repo.UpdateSettingDTO{}
	if err := copier.Copy(&params, setting); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying setting to params: %w",
			err,
		)
	}

	if err := uc.val.Validate(params); err != nil {
		return entity.Setting{}, err
	}

	updatedSetting, err := uc.settingRepo.UpdateSetting(setting.ID, params)
	if err != nil {
		return entity.Setting{}, fmt.Errorf("error updating setting: %w", err)
	}

	return updatedSetting, nil
}

func (uc *UpsertUserSettingUseCase) createSetting(
	dto UpsertUserSettingDTO,
) (entity.Setting, error) {
	params := repo.CreateSettingDTO{}
	if err := copier.Copy(&params, dto); err != nil {
		return entity.Setting{}, fmt.Errorf(
			"error copying dto to params: %w",
			err,
		)
	}

	if err := uc.val.Validate(params); err != nil {
		return entity.Setting{}, err
	}

	setting, err := uc.settingRepo.CreateSetting(params)
	if err != nil {
		return entity.Setting{}, fmt.Errorf("error creating setting: %w", err)
	}

	return setting, nil
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
