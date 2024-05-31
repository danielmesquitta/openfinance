package middleware

import (
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/logger"
)

type Middleware struct {
	log       *logger.Logger
	jwtIssuer *jwt.JWTIssuer
}

func NewMiddleware(
	log *logger.Logger,
	jwtIssuer *jwt.JWTIssuer,
) *Middleware {
	return &Middleware{
		log:       log,
		jwtIssuer: jwtIssuer,
	}
}
