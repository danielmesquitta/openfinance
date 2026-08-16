package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	root "github.com/danielmesquitta/openfinance-to-sheets"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/pkg/validator"
	"github.com/spf13/viper"
)

// EnvFileData is the data for the .env file.
type EnvFileData struct {
	OpenAIToken             string `json:"open_ai_token"             mapstructure:"OPEN_AI_TOKEN"             validate:"required"`
	MaxConcurrentOperations int    `json:"max_concurrent_operations" mapstructure:"MAX_CONCURRENT_OPERATIONS" validate:"required,gte=1"`
}

// IngestProfilesFileData is the data for the ingest_profiles.json file.
type IngestProfilesFileData struct {
	IngestProfiles []entity.IngestProfile `json:"ingest_profiles" validate:"required,min=1,unique=ID,dive"`
}

// Env is the environment variables.
type Env struct {
	EnvFileData
	IngestProfilesFileData

	IngestSettings entity.IngestSettings

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

	if err := e.loadDataFromIngestProfilesFile(); err != nil {
		return fmt.Errorf("failed to load data from ingest profiles file: %w", err)
	}

	if err := e.validateIngestProfilesFile(); err != nil {
		return fmt.Errorf("failed to validate ingest profiles file: %w", err)
	}

	return e.loadIngestSettings()
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

func (e *Env) loadDataFromIngestProfilesFile() error {
	if err := e.loadIngestProfiles(); err != nil {
		return fmt.Errorf("failed to load ingest profiles: %w", err)
	}

	return nil
}

func (e *Env) loadIngestSettings() error {
	settings, err := entity.NewIngestSettings(e.IngestProfiles)
	if err != nil {
		return fmt.Errorf("invalid domain settings: %w", err)
	}

	e.IngestSettings = settings

	return nil
}

func (e *Env) loadIngestProfiles() (err error) {
	ingestProfilesData, err := root.IngestProfilesFile.ReadFile("config/ingest_profiles.json")
	if err != nil {
		return fmt.Errorf("failed to read ingest profiles file: %w", err)
	}

	if err = json.Unmarshal(ingestProfilesData, &e.IngestProfiles); err != nil {
		return fmt.Errorf("failed to unmarshal ingest profiles file: %w", err)
	}

	return nil
}

func (e *Env) validateEnvFile() error {
	if err := e.val.Validate(e.EnvFileData); err != nil {
		return fmt.Errorf("failed to validate env file: %w", err)
	}

	return nil
}

func (e *Env) validateIngestProfilesFile() error {
	if err := e.val.Validate(e.IngestProfilesFileData); err != nil {
		return fmt.Errorf("failed to validate ingest profiles file: %w", err)
	}

	return nil
}
