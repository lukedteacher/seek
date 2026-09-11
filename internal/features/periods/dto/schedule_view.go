package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
	"time"
)

type PeriodScheduleView struct {
	models.Period
	Row       int `json:"row"`
	Column    int `json:"column"`
	Educators []educatorDTO.EducatorView
	Students  []studentDTO.StudentView
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
					Period: period,
					Row:    row,
					Column: day.Column(),
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
