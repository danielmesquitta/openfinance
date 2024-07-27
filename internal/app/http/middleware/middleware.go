package middleware

import (
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
)

type Middleware struct {
	env       *config.Env
	jwtIssuer *jwt.JWTIssuer
}

func NewMiddleware(
	env *config.Env,
	jwtIssuer *jwt.JWTIssuer,
) *Middleware {
	return &Middleware{
		env:       env,
		jwtIssuer: jwtIssuer,
	}
}
