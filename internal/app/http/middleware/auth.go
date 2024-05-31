package middleware

import (
	"strings"

	"github.com/danielmesquitta/openfinance/internal/app/http/dto"
	"github.com/gofiber/fiber/v2"
)

const UserIDKey = "userID"

func (m *Middleware) EnsureAuthenticated(ctx *fiber.Ctx) error {
	authorization := ctx.Get("Authorization")

	if authorization == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponseDTO{
			Message: "unauthorized",
		})
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponseDTO{
			Message: "unauthorized",
		})
	}

	accessToken := strings.TrimSpace(
		strings.TrimPrefix(authorization, "Bearer "),
	)

	userID, err := m.j.ParseToken(accessToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponseDTO{
			Message: "unauthorized",
		})
	}

	ctx.Locals(UserIDKey, userID)

	return ctx.Next()
}
