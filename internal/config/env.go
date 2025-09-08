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
	OpenAIToken string `mapstructure:"OPEN_AI_TOKEN"   json:"open_ai_token"   validate:"required"`
}

// JSONFileData is the data for the users.json file.
type JSONFileData struct {
	Users []entity.User `json:"users" validate:"required"`
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
		return err
	}

	if err := e.validateEnvFile(); err != nil {
		return err
	}

	if err := e.loadDataFromJSON(); err != nil {
		return err
	}

	if err := e.validateJSONFile(); err != nil {
		return err
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
	users := []entity.User{}

	data, err := root.UsersFile.ReadFile("users.json")
	if err != nil {
		return fmt.Errorf("failed to read users file: %w", err)
	}

	if err = json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to unmarshal users file: %w", err)
	}

	e.Users = users

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
