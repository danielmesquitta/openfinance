package usecase

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
)

type SyncAll struct {
	val            *validator.Validator
	env            *config.Env
	syncOneUseCase *SyncOne
}

func NewSyncAll(
	val *validator.Validator,
	env *config.Env,
	syncOneUseCase *SyncOne,
) *SyncAll {
	return &SyncAll{
		val:            val,
		env:            env,
		syncOneUseCase: syncOneUseCase,
	}
}

func (sa *SyncAll) Execute(
	ctx context.Context,
	dto SyncDTO,
) error {
	if err := sa.val.Validate(dto); err != nil {
		return fmt.Errorf("failed to validate dto: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	for _, user := range sa.env.Users {
		g.Go(func() error {
			return sa.syncOneUseCase.Execute(
				gCtx,
				user.ID,
				dto,
			)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to wait for sync all: %w", err)
	}

	return nil
}
