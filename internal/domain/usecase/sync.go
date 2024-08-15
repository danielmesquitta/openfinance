package usecase

import (
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type SyncDTO struct {
	StartDate string `validate:"datetime=2006-01-02T15:04:05Z07:00" json:"start_date,omitempty"`
	EndDate   string `validate:"datetime=2006-01-02T15:04:05Z07:00" json:"end_date,omitempty"`
}

// setDefaultValues sets default values for the SyncDTO
// if StartDate is not provided, it will be set to the first day of the previous month
// if EndDate is not provided, it will be set to the last day of the previous month
func setDefaultValues(
	dto *SyncDTO,
) {
	now := time.Now()
	startOfMonth := time.Date(
		now.Year(),
		now.Month()-1,
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
	if dto.StartDate == "" {
		dto.StartDate = startOfMonth.Format(time.RFC3339)
	}
	if dto.EndDate == "" {
		dto.EndDate = endOfMonth.Format(time.RFC3339)
	}
}

// parseDates parses the StartDate and EndDate from the SyncDTO
// and returns them as time.Time
func parseDates(
	dto SyncDTO,
) (startDate time.Time, endDate time.Time, err error) {
	invalidDateErr := entity.ErrValidation
	invalidDateErr.Message = "invalid date"

	startDate, err = time.Parse(time.RFC3339, dto.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, invalidDateErr
	}

	endDate, err = time.Parse(time.RFC3339, dto.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, invalidDateErr
	}

	return startDate, endDate, nil
}
