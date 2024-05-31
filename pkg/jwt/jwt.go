package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/golang-jwt/jwt"
)

type JWTIssuer struct {
	env *config.Env
}

func NewJWTIssuer(
	env *config.Env,
) *JWTIssuer {
	return &JWTIssuer{
		env: env,
	}
}

func (j *JWTIssuer) NewAccessToken(
	userID string,
) (accessToken string, expiresAt int64, err error) {
	expiresAt = time.Now().Add(time.Hour * 24).Unix()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    userID,
		ExpiresAt: expiresAt,
	})

	accessToken, err = claims.SignedString([]byte(j.env.JWTSecret))
	if err != nil {
		return "", 0, fmt.Errorf("error signing token: %w", err)
	}

	return accessToken, expiresAt, nil
}

type Claims struct {
	Issuer    string `json:"iss"`
	ExpiresAt int64  `json:"exp"`
}

func (j *JWTIssuer) ParseToken(
	accessToken string,
) (userID string, err error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(j.env.JWTSecret), nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error parsing token: %w", err)
	}

	bytes, err := json.Marshal(token.Claims)
	if err != nil {
		return "", fmt.Errorf("error marshalling claims: %w", err)
	}

	claims := Claims{}
	if err := json.Unmarshal(bytes, &claims); err != nil {
		return "", fmt.Errorf("error unmarshalling claims: %w", err)
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return "", entity.ErrTokenExpired
	}

	return claims.Issuer, nil
}
