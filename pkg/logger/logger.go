package logger

import (
	"github.com/danielmesquitta/openfinance/internal/config"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(env *config.Env) *Logger {
	if env.Environment == config.DevelopmentEnv {
		return &Logger{
			SugaredLogger: zap.Must(zap.NewDevelopment()).Sugar(),
		}
	}

	return &Logger{
		SugaredLogger: zap.Must(zap.NewProduction()).Sugar(),
	}
}
