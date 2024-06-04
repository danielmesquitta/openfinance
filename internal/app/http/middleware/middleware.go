package middleware

import (
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/logger"
)

type Middleware struct {
	env       *config.Env
	log       *logger.Logger
	jwtIssuer *jwt.JWTIssuer
}

func NewMiddleware(
	env *config.Env,
	log *logger.Logger,
	jwtIssuer *jwt.JWTIssuer,
) *Middleware {
	return &Middleware{
		env:       env,
		log:       log,
		jwtIssuer: jwtIssuer,
	}
}
