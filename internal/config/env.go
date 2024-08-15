package config

import (
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/spf13/viper"
)

type Environment string

const (
	DevelopmentEnv Environment = "development"
	TestEnv        Environment = "test"
	ProductionEnv  Environment = "production"
)

type Env struct {
	val *validator.Validator

	Environment             Environment `mapstructure:"ENVIRONMENT"`
	Port                    string      `mapstructure:"PORT"`
	APIURL                  string      `mapstructure:"API_URL"`
	GoogleOAUTHClientID     string      `mapstructure:"GOOGLE_OAUTH_CLIENT_ID"     validate:"required"`
	GoogleOAUTHClientSecret string      `mapstructure:"GOOGLE_OAUTH_CLIENT_SECRET" validate:"required"`
	BasicAuthUsername       string      `mapstructure:"BASIC_AUTH_USERNAME"        validate:"required"`
	BasicAuthPassword       string      `mapstructure:"BASIC_AUTH_PASSWORD"        validate:"required"`
	JWTSecret               string      `mapstructure:"JWT_SECRET"                 validate:"required"`
	HashSecret              string      `mapstructure:"HASH_SECRET"                validate:"required"`
	OpenAIAPIToken          string      `mapstructure:"OPEN_AI_API_TOKEN"          validate:"required"`

	// Optional (not required for AWS lambda)
	DBConnection string `mapstructure:"DB_CONNECTION"`
}

func (e *Env) validate() error {
	if err := e.val.Validate(e); err != nil {
		return err
	}
	if e.Environment == "" {
		e.Environment = DevelopmentEnv
	}
	if e.Port == "" {
		e.Port = "8080"
	}
	if e.APIURL == "" {
		e.APIURL = "http://localhost:" + e.Port
	}
	return nil
}

func LoadEnv(val *validator.Validator) *Env {
	env := &Env{
		val: val,
	}

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&env); err != nil {
		panic(err)
	}

	if err := env.validate(); err != nil {
		panic(err)
	}

	return env
}
