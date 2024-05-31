package middleware

import (
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/logger"
)

type Middleware struct {
	l *logger.Logger
	j *jwt.JWTIssuer
}

func NewMiddleware(
	l *logger.Logger,
	j *jwt.JWTIssuer,
) *Middleware {
	return &Middleware{
		l: l,
		j: j,
	}
}
