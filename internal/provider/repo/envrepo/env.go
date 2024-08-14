package envrepo

import (
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/spf13/viper"
)

type Env struct {
	val *validator.Validator

	NotionToken           string `mapstructure:"NOTION_TOKEN"             validate:"required"`
	NotionPageID          string `mapstructure:"NOTION_PAGE_ID"           validate:"required"`
	MeuPluggyClientID     string `mapstructure:"MEU_PLUGGY_CLIENT_ID"     validate:"required"`
	MeuPluggyClientSecret string `mapstructure:"MEU_PLUGGY_CLIENT_SECRET" validate:"required"`
	MeuPluggyAccountIDs   string `mapstructure:"MEU_PLUGGY_ACCOUNT_IDS"   validate:"required"`
}

func (e *Env) validate() error {
	if err := e.val.Validate(e); err != nil {
		return err
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
