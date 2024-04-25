package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Env struct {
	Port                  string   `mapstructure:"PORT"`
	NotionToken           string   `mapstructure:"NOTION_TOKEN"`
	NotionPageID          string   `mapstructure:"NOTION_PAGE_ID"`
	MeuPluggyClientID     string   `mapstructure:"MEU_PLUGGY_CLIENT_ID"`
	MeuPluggyClientSecret string   `mapstructure:"MEU_PLUGGY_CLIENT_SECRET"`
	MeuPluggyAccountIDs   []string `mapstructure:"MEU_PLUGGY_ACCOUNT_IDS"`
	MeuPluggyToken        string
}

func (e *Env) validate() error {
	errs := []string{}
	if e.Port == "" {
		e.Port = "8080"
	}
	if e.NotionToken == "" {
		errs = append(errs, "NOTION_TOKEN (string) is not set")
	}
	if e.NotionPageID == "" {
		errs = append(errs, "NOTION_PAGE_ID (string) is not set")
	}
	if e.MeuPluggyClientID == "" {
		errs = append(errs, "MEU_PLUGGY_CLIENT_ID (string) is not set")
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
