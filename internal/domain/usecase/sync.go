package usecase

import (
	"fmt"
	"time"
)

type SyncDTO struct {
	StartDate string `json:"start_date,omitempty" validate:"datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `json:"end_date,omitempty"   validate:"datetime=2006-01-02T15:04:05Z07:00"`
}

// parseDates parses the StartDate and EndDate from the SyncDTO
// and returns them as time.Time.
func parseDates(
	dto SyncDTO,
) (startDate time.Time, endDate time.Time, err error) {
	startDate, err = time.Parse(time.RFC3339, dto.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	endDate, err = time.Parse(time.RFC3339, dto.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}

	return startDate, endDate, nil
}
