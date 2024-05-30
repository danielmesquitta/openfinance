package middleware

import (
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/app/http/dto"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type Middleware struct {
	log *logger.Logger
}

func NewMiddleware(
	log *logger.Logger,
) *Middleware {
	return &Middleware{
		log: log,
	}
}

func (m *Middleware) ErrorHandler(ctx *fiber.Ctx, err error) error {
	if appErr, ok := err.(*entity.AppError); ok {
		return ctx.Status(appErr.HTTPStatusCode).
			JSON(dto.ErrorResponseDTO{
				Message: appErr.Message,
			})
	}

	if fiberErr, ok := err.(*fiber.Error); ok && fiberErr.Code < 500 &&
		fiberErr.Code >= 300 {
		return ctx.Status(fiberErr.Code).
			JSON(dto.ErrorResponseDTO{
				Message: fiberErr.Message,
			})
	}

	m.log.Error(
		"internal server error",
		"error",
		err,
		"body",
		string(ctx.Body()),
		"query",
		ctx.Queries(),
		"params",
		ctx.AllParams(),
	)

	return ctx.Status(http.StatusInternalServerError).
		JSON(dto.ErrorResponseDTO{
			Message: "Internal server error",
		})
}
