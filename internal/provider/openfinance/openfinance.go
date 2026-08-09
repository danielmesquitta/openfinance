package openfinance

import (
	"context"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type APIProvider interface {
	ListTransactionsBySyncProfileID(
		ctx context.Context,
		syncProfileID string,
		from, to time.Time,
	) ([]entity.Transaction, error)
}
