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
	p models.Period,
) []PeriodScheduleView {
	row := timeToRow(p.StartTime, 479)
	views := make([]PeriodScheduleView, 0)

	for _, day := range sharedmodels.Days {
		if int(p.DaysBitmask)&day.Bit() != 0 {
			views = append(views, PeriodScheduleView{
				ID:          p.ID,
				Title:       p.Title,
				ServiceType: p.ServiceType,
				StartTime:   p.StartTime,
				EndTime:     p.EndTime,
				Duration:    p.Duration,
				Row:         row,
				Column:      day.Column(),
			})
		}
	}
	return views
}

func timeToRow(t sharedmodels.TimeOnly, offset int) int {
	tt := time.Time(t)
	totalMinutes := tt.Hour()*60 + tt.Minute()
	return totalMinutes - offset
}
