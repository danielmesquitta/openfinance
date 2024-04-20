package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Env struct {
	NotionToken              string `mapstructure:"NOTION_TOKEN"`
	NotionUsername           string `mapstructure:"NOTION_USERNAME"`
	NotionPageID             string `mapstructure:"NOTION_PAGE_ID"`
	MeuPluggyClientID        string `mapstructure:"MEU_PLUGGY_CLIENT_ID"`
	MeuPluggyClientSecret    string `mapstructure:"MEU_PLUGGY_CLIENT_SECRET"`
	MeuPluggyBankAccountID   string `mapstructure:"MEU_PLUGGY_BANK_ACCOUNT_ID"`
	MeuPluggyCreditAccountID string `mapstructure:"MEU_PLUGGY_CREDIT_ACCOUNT_ID"`
	MeuPluggyToken           string
}

func (e *Env) validate() error {
	errs := []string{}
	if e.NotionToken == "" {
		errs = append(errs, "NOTION_TOKEN (string) is not set")
	}
	if e.NotionUsername == "" {
		errs = append(errs, "NOTION_USERNAME (string) is not set")
	}
	if e.NotionPageID == "" {
		errs = append(errs, "NOTION_PAGE_ID (string) is not set")
	}
	if e.MeuPluggyClientID == "" {
		errs = append(errs, "MEU_PLUGGY_CLIENT_ID (string) is not set")
	}
	if e.MeuPluggyBankAccountID == "" {
		errs = append(errs, "MEU_PLUGGY_BANK_ACCOUNT_ID (string) is not set")
	}
	if e.MeuPluggyCreditAccountID == "" {
		errs = append(errs, "MEU_PLUGGY_CREDIT_ACCOUNT_ID (string) is not set")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}

func LoadEnv() *Env {
	env := &Env{}

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
