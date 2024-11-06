package usecase

import (
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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

	fmt.Println("dto", dto.StartDate, dto.EndDate)

	if err := sa.val.Validate(dto); err != nil {
		return err
	}

	jobsCount := len(sa.env.Users)
	errCh := make(chan error)

	for _, user := range sa.env.Users {
		go func() {
			err := sa.syncOneUseCase.Execute(
				user.ID,
				dto,
			)
			errCh <- err
		}()
	}

	var err error
	for i := 0; i < jobsCount; i++ {
		err = errors.Join(err, <-errCh)
	}

	close(errCh)

	if err != nil {
		return entity.NewErr(err)
	}

	return nil
}
