package config

import (
	"cmp"
	"encoding/json"
	"io"
	"os"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/spf13/viper"
)

type Environment string

const (
	DevelopmentEnv Environment = "development"
	TestEnv        Environment = "test"
	ProductionEnv  Environment = "production"
)

type EnvFileData struct {
	Environment   Environment `mapstructure:"ENVIRONMENT"     json:"environment"`
	OpenAIToken   string      `mapstructure:"OPEN_AI_TOKEN"   json:"open_ai_token"   validate:"required"`
	UsersFilePath string      `mapstructure:"USERS_FILE_PATH" json:"users_file_path" validate:"required"`
}

type JSONFileData struct {
	Users []entity.User `json:"users" validate:"required"`
}

type Env struct {
	val *validator.Validator

	EnvFileData
	JSONFileData
}

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

	if err := e.validateEnvFile(e.EnvFileData); err != nil {
		return err
	}

	if err := e.loadDataFromJSON(); err != nil {
		return err
	}

	if err := e.validateJSONFile(e.JSONFileData); err != nil {
		return err
	}

	return nil
}

func (e *Env) loadDataFromEnvFile() error {
	envFilepath := cmp.Or(os.Getenv("ENV_FILEPATH"), ".env")

	viper.SetConfigFile(envFilepath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(e); err != nil {
		return err
	}

	return nil
}

func (e *Env) loadDataFromJSON() error {
	users := []entity.User{}

	file, err := os.Open(e.UsersFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &users); err != nil {
		return err
	}

	e.Users = users

	return nil
}

func (e *Env) validateEnvFile(data EnvFileData) error {
	if err := e.val.Validate(data); err != nil {
		return err
	}
	if e.Environment == "" {
		e.Environment = DevelopmentEnv
	}
	return nil
}

func (e *Env) validateJSONFile(data JSONFileData) error {
	if err := e.val.Validate(data); err != nil {
		return err
	}
	if e.Environment == "" {
		e.Environment = DevelopmentEnv
	}
	return nil
}
