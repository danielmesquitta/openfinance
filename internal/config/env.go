package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	root "github.com/danielmesquitta/openfinance"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/spf13/viper"
)

// EnvFileData is the data for the .env file.
type EnvFileData struct {
	OpenAIToken             string `json:"open_ai_token"             mapstructure:"OPEN_AI_TOKEN"             validate:"required"`
	MaxConcurrentOperations int    `json:"max_concurrent_operations" mapstructure:"MAX_CONCURRENT_OPERATIONS" validate:"required,gte=1"`
}

// SyncProfilesFileData is the data for the sync_profiles.json file.
type SyncProfilesFileData struct {
	SyncProfiles []entity.SyncProfile `json:"sync_profiles" validate:"required"`
}

// Env is the environment variables.
type Env struct {
	EnvFileData
	SyncProfilesFileData

	SyncSettings entity.SyncSettings

	val *validator.Validator
}

// NewEnv creates a new Env.
func NewEnv(val *validator.Validator) (*Env, error) {
	e := &Env{
		val: val,
	}

	if err := e.loadEnv(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Env) loadEnv() error {
	if err := e.loadDataFromEnvFile(); err != nil {
		return fmt.Errorf("failed to load data from env file: %w", err)
	}

	if err := e.validateEnvFile(); err != nil {
		return fmt.Errorf("failed to validate env file: %w", err)
	}

	if err := e.loadDataFromSyncProfilesFile(); err != nil {
		return fmt.Errorf("failed to load data from sync profiles file: %w", err)
	}

	if err := e.validateSyncProfilesFile(); err != nil {
		return fmt.Errorf("failed to validate sync profiles file: %w", err)
	}

	return nil
}

func (e *Env) loadDataFromEnvFile() error {
	envFile, err := root.EnvFile.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	viper.SetConfigType("env")

	if err := viper.ReadConfig(bytes.NewBuffer(envFile)); err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&e.EnvFileData); err != nil {
		return fmt.Errorf("failed to unmarshal env file: %w", err)
	}

	return nil
}

func (e *Env) loadDataFromSyncProfilesFile() error {
	if err := e.loadSyncProfiles(); err != nil {
		return fmt.Errorf("failed to load sync profiles: %w", err)
	}

	settings, err := entity.NewSyncSettings(e.SyncProfiles)
	if err != nil {
		return fmt.Errorf("invalid domain settings: %w", err)
	}

	e.SyncSettings = settings

	return nil
}

func (e *Env) loadSyncProfiles() (err error) {
	syncProfilesData, err := root.SyncProfilesFile.ReadFile("config/sync_profiles.json")
	if err != nil {
		return fmt.Errorf("failed to read sync profiles file: %w", err)
	}

	if err = json.Unmarshal(syncProfilesData, &e.SyncProfiles); err != nil {
		return fmt.Errorf("failed to unmarshal sync profiles file: %w", err)
	}

	return nil
}

func (e *Env) validateEnvFile() error {
	if err := e.val.Validate(e.EnvFileData); err != nil {
		return fmt.Errorf("failed to validate env file: %w", err)
	}

	return nil
}

func (e *Env) validateSyncProfilesFile() error {
	if err := e.val.Validate(e.SyncProfilesFileData); err != nil {
		return fmt.Errorf("failed to validate sync profiles file: %w", err)
	}

	return nil
}
