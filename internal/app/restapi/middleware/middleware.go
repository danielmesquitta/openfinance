package middleware

import (
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
)

type Middleware struct {
	env    *config.Env
	Issuer *jwt.Issuer
}

func NewMiddleware(
	env *config.Env,
	Issuer *jwt.Issuer,
) *Middleware {
	return &Middleware{
		env:    env,
		Issuer: Issuer,
	}
}
