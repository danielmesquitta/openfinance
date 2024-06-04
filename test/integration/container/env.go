package container

import (
	"fmt"
	"math/rand"
	"path/filepath"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/spf13/viper"
)

func generateRandomPort() int {
	minPort := 1000
	maxPort := 9999
	return rand.Intn(maxPort-minPort+1) + minPort
}

func loadTestEnv(dbConnURL string) *config.Env {
	env := &config.Env{}

	viper.SetConfigFile(filepath.Join(CWD, ".env"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&env); err != nil {
		panic(err)
	}

	port := fmt.Sprintf("%d", generateRandomPort())

	env.Environment = config.TestEnv
	env.Port = port
	env.DBConnection = dbConnURL
	env.ApiURL = "http://localhost:" + port

	return env
}
