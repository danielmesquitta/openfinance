package logger

import (
	"github.com/danielmesquitta/openfinance/internal/config"
	"go.uber.org/zap"
)

type Logger struct {
	env *config.Env
	log *zap.SugaredLogger
}

func NewLogger(env *config.Env) *Logger {
	if env.Environment == config.DevelopmentEnv {
		return &Logger{
			env: env,
			log: zap.Must(zap.NewDevelopment()).Sugar(),
		}
	}

	return &Logger{
		env: env,
		log: zap.Must(zap.NewProduction()).Sugar(),
	}
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.log.Infow(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.log.Errorw(msg, keysAndValues...)
}
