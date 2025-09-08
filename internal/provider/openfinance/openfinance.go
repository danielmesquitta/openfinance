package openfinance

import (
	"context"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type APIProvider interface {
	ListTransactionsByUserID(
		ctx context.Context,
		userID string,
		from, to time.Time,
	) ([]entity.Transaction, error)
}
