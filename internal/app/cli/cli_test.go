package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"

	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase/mockusecase"
)

func testCommand(startDate, endDate time.Time) *cobra.Command {
	command := &cobra.Command{}
	command.Flags().Int(monthFlag, int(startDate.Month()), "")
	command.Flags().Int(yearFlag, startDate.Year(), "")
	command.Flags().Time(startDateFlag, startDate, timeFormats, "")
	command.Flags().Time(endDateFlag, endDate, timeFormats, "")

	return command
}

func TestExecuteSyncPropagatesSyncError(t *testing.T) {
	t.Parallel()

	startDate := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	wantErr := errors.New("sync failed")
	syncUseCase := mockusecase.NewMockSyncExecutor(t)
	syncUseCase.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(input usecase.SyncInput) bool {
			return input.StartDate.Equal(startDate) && input.EndDate.Equal(endDate)
		})).
		Return(wantErr).
		Once()

	err := executeSync(testCommand(startDate, endDate), syncUseCase)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeSync() error = %v", err)
	}
}

func TestExecuteSyncUsesMonthAndYearFlags(t *testing.T) {
	t.Parallel()

	defaultStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.Local)
	defaultEnd := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.Local)
	command := testCommand(defaultStart, defaultEnd)
	if err := command.Flags().Set(monthFlag, "2"); err != nil {
		t.Fatalf("set month flag: %v", err)
	}
	if err := command.Flags().Set(yearFlag, "2025"); err != nil {
		t.Fatalf("set year flag: %v", err)
	}

	wantStart := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.Local)
	wantEnd := wantStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	syncUseCase := mockusecase.NewMockSyncExecutor(t)
	syncUseCase.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(input usecase.SyncInput) bool {
			return input.StartDate.Equal(wantStart) && input.EndDate.Equal(wantEnd)
		})).
		Return(nil).
		Once()

	if err := executeSync(command, syncUseCase); err != nil {
		t.Fatalf("executeSync() error = %v", err)
	}
}
