package openfinance

import (
	"context"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type APIProvider interface {
	ListTransactionsByIngestProfileID(
		ctx context.Context,
		ingestProfileID string,
		from, to time.Time,
	) ([]entity.Transaction, error)
}
