package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

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

// JSONFileData is the data for the users.json file.
type JSONFileData struct {
	Users            []entity.User                    `json:"users"              validate:"required"`
	ColorsByCategory map[entity.Category]entity.Color `json:"colors_by_category" validate:"required"`
	Categories       []entity.Category                `json:"categories"         validate:"required"`
	Mappings         map[string]entity.Category       `json:"mappings"           validate:"required"`
}

// Env is the environment variables.
type Env struct {
	EnvFileData
	JSONFileData

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

	if err := e.loadDataFromJSON(); err != nil {
		return fmt.Errorf("failed to load data from users file: %w", err)
	}

	if err := e.validateJSONFile(); err != nil {
		return fmt.Errorf("failed to validate users file: %w", err)
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

func (e *Env) loadDataFromJSON() error {
	if err := e.loadCategories(); err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	if err := e.loadMappings(); err != nil {
		return fmt.Errorf("failed to load mappings: %w", err)
	}

	if err := e.loadUsers(); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	settings, err := entity.NewSyncSettings(e.Users, e.ColorsByCategory, e.Mappings)
	if err != nil {
		return fmt.Errorf("invalid domain settings: %w", err)
	}

	e.SyncSettings = settings

	return nil
}

func (e *Env) loadCategories() error {
	categoriesData, err := root.Config.ReadFile("config/categories.json")
	if err != nil {
		return fmt.Errorf("failed to read categories file: %w", err)
	}

	if err = json.Unmarshal(categoriesData, &e.ColorsByCategory); err != nil {
		return fmt.Errorf("failed to unmarshal categories file: %w", err)
	}

	categories := make([]entity.Category, 0, len(e.ColorsByCategory))
	for category := range e.ColorsByCategory {
		categories = append(categories, category)
	}
	slices.Sort(categories)
	e.Categories = categories

	return nil
}

func (e *Env) loadMappings() error {
	mappingsData, err := root.Config.ReadFile("config/mappings.json")
	if err != nil {
		return fmt.Errorf("failed to read mappings file: %w", err)
	}

	if err = json.Unmarshal(mappingsData, &e.Mappings); err != nil {
		return fmt.Errorf("failed to unmarshal mappings file: %w", err)
	}

	return nil
}

func (e *Env) loadUsers() (err error) {
	usersData, err := root.Config.ReadFile("config/users.json")
	if err != nil {
		return fmt.Errorf("failed to read users file: %w", err)
	}

	if err = json.Unmarshal(usersData, &e.Users); err != nil {
		return fmt.Errorf("failed to unmarshal users file: %w", err)
	}

	return nil
}

func (e *Env) validateEnvFile() error {
	if err := e.val.Validate(e.EnvFileData); err != nil {
		return fmt.Errorf("failed to validate env file: %w", err)
	}

	return nil
}

func (e *Env) validateJSONFile() error {
	if err := e.val.Validate(e.JSONFileData); err != nil {
		return fmt.Errorf("failed to validate users file: %w", err)
	}

	return nil
}
