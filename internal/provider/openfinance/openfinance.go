package openfinance

import (
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type APIProvider interface {
	ListTransactionsByUserID(
		userID string,
		from, to time.Time,
	) ([]entity.Transaction, error)
}
