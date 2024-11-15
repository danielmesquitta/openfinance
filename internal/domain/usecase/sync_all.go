package usecase

import (
	"github.com/sourcegraph/conc/iter"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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
	dto SyncDTO,
) error {
	setDefaultValues(&dto)

	if err := sa.val.Validate(dto); err != nil {
		return err
	}

	_, err := iter.MapErr(
		sa.env.Users,
		func(user *entity.User) (*struct{}, error) {
			err := sa.syncOneUseCase.Execute(
				user.ID,
				dto,
			)
			return nil, err
		},
	)

	if err != nil {
		return errs.New(err)
	}

	return nil
}
