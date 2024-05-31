package middleware

import (
	"errors"
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/app/http/dto"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) ErrorHandler(ctx *fiber.Ctx, err error) error {
	appErr := entity.AppError{}
	if ok := errors.As(err, &appErr); ok {
		return ctx.Status(appErr.HTTPStatusCode).
			JSON(dto.ErrorResponseDTO{
				Message: appErr.Message,
			})
	}

	m.l.Errorw(
		"internal server error",
		"error",
		err,
		"url",
		ctx.OriginalURL(),
		"body",
		string(ctx.Body()),
		"query",
		ctx.Queries(),
		"params",
		ctx.AllParams(),
	)

	if fiberErr, ok := err.(*fiber.Error); ok {
		return ctx.Status(fiberErr.Code).
			JSON(dto.ErrorResponseDTO{
				Message: fiberErr.Message,
			})
	}

	return ctx.Status(http.StatusInternalServerError).
		JSON(dto.ErrorResponseDTO{
			Message: "Internal server error",
		})
}
