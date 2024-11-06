package cli

import (
	"time"

	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

func Run() error {
	syncAllUseCase := app.NewSyncAllUseCase()

	startOfMonth := time.Date(
		time.Now().Year(),
		time.Now().Month(),
		1,
		0,
		0,
		0,
		0,
		time.Local,
	)
	endOfMonth := startOfMonth.AddDate(
		0,
		1,
		-1,
	)

	err := syncAllUseCase.Execute(usecase.SyncDTO{
		StartDate: startOfMonth.Format(time.RFC3339),
		EndDate:   endOfMonth.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	return nil
}
