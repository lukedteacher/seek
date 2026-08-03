package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"
	"seek/internal/features/students/dto"
	"time"
)

type PeriodScheduleView struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	EndTime     sharedmodels.TimeOnly    `json:"end_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	Row         int                      `json:"row"`
	Column      int                      `json:"column"`
	Students    []dto.StudentView
}

func NewPeriodScheduleViews(
	periods ...models.Period,
) []PeriodScheduleView {
	periodScheduleViews := make([]PeriodScheduleView, 0)
	for _, period := range periods {
		row := timeToRow(period.StartTime, 479)

		for _, day := range sharedmodels.Days {
			if int(period.DaysBitmask)&day.Bit() != 0 {
				periodScheduleViews = append(periodScheduleViews, PeriodScheduleView{
					ID:          period.ID,
					Title:       period.Title,
					ServiceType: period.ServiceType,
					StartTime:   period.StartTime,
					EndTime:     period.EndTime,
					Duration:    period.Duration,
					Row:         row,
					Column:      day.Column(),
				})
			}
		}
	}
	return periodScheduleViews
}

func timeToRow(t sharedmodels.TimeOnly, offset int) int {
	tt := time.Time(t)
	totalMinutes := tt.Hour()*60 + tt.Minute()
	return totalMinutes - offset
}
