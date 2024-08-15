package cli

import (
	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

func Run() error {
	syncAllUseCase := app.NewSyncAllUseCase()

	err := syncAllUseCase.Execute(usecase.SyncDTO{})
	if err != nil {
		return err
	}

	return nil
}
