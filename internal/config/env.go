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
	OpenAIToken string `mapstructure:"OPEN_AI_TOKEN" json:"open_ai_token" validate:"required"`
}

// JSONFileData is the data for the users.json file.
type JSONFileData struct {
	Users            []entity.User                    `json:"users"              validate:"required"`
	ColorsByCategory map[entity.Category]entity.Color `json:"colors_by_category" validate:"required"`
	Categories       []entity.Category                `json:"categories"         validate:"required"`
	JSONCategories   []byte                           `json:"json_categories"    validate:"required"`
	Mappings         map[string]entity.Category       `json:"mappings"           validate:"required"`
	JSONMappings     []byte                           `json:"json_mappings"      validate:"required"`
}

// Env is the environment variables.
type Env struct {
	val *validator.Validator

	EnvFileData
	JSONFileData
}

// NewEnv creates a new Env.
func NewEnv(val *validator.Validator) *Env {
	e := &Env{
		val: val,
	}

	if err := e.loadEnv(); err != nil {
		panic(err)
	}

	return e
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
	uniqueCategories, err := e.loadCategories()
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	if err := e.loadMappings(uniqueCategories); err != nil {
		return fmt.Errorf("failed to load mappings: %w", err)
	}

	if err := e.loadUsers(); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	return nil
}

func (e *Env) loadCategories() (uniqueCategories map[entity.Category]struct{}, err error) {
	categoriesData, err := root.Config.ReadFile("config/categories.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read categories file: %w", err)
	}

	if err = json.Unmarshal(categoriesData, &e.ColorsByCategory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories file: %w", err)
	}

	uniqueCategories = map[entity.Category]struct{}{}
	uniqueColors := map[entity.Color]struct{}{}
	categories := make([]entity.Category, 0, len(e.ColorsByCategory))
	for category, color := range e.ColorsByCategory {
		if _, ok := uniqueCategories[category]; ok {
			return nil, fmt.Errorf("category %s is not unique", category)
		}
		if _, ok := entity.ColorsMap[color]; !ok {
			return nil, fmt.Errorf("color %s is not valid", color)
		}
		if _, ok := uniqueColors[color]; ok {
			return nil, fmt.Errorf("color %s is not unique", color)
		}
		uniqueCategories[category] = struct{}{}
		uniqueColors[color] = struct{}{}
		categories = append(categories, category)
	}

	categoriesData, err = json.Marshal(categories)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal categories: %w", err)
	}

	e.JSONCategories = categoriesData
	e.Categories = categories

	return uniqueCategories, nil
}

func (e *Env) loadMappings(uniqueCategories map[entity.Category]struct{}) (err error) {
	mappingsData, err := root.Config.ReadFile("config/mappings.json")
	if err != nil {
		return fmt.Errorf("failed to read mappings file: %w", err)
	}

	if err = json.Unmarshal(mappingsData, &e.Mappings); err != nil {
		return fmt.Errorf("failed to unmarshal mappings file: %w", err)
	}

	for _, category := range e.Mappings {
		if _, ok := uniqueCategories[category]; !ok {
			return fmt.Errorf("category %s is not found in categories", category)
		}
	}

	e.JSONMappings = mappingsData

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
