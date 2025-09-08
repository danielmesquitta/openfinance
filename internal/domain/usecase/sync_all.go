package usecase

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
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
	setDefaultValues(&dto)
	if err := sa.val.Validate(dto); err != nil {
		return err
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
		return errs.New(err)
	}

	return nil
}
