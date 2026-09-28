package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	periodDTO "seek/internal/features/periods/dto"
	"seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
	"time"
)

type GridPosition struct {
	Row    int
	Column int
}

func NewGridPosition(time sharedmodels.TimeOnly, day sharedmodels.Day) GridPosition {
	return GridPosition{
		Row:    timeToRow(time, 479),
		Column: dayToColumn(day),
	}
}

type SchedulePeriodView struct {
	Period    periodDTO.PeriodView
	Position  GridPosition
	Educators []educatorDTO.EducatorView
	Students  []studentDTO.StudentView
}

// iterates thru periods and returns one item per day per period
func NewSchedulePeriodViews(
	periods ...models.Period,
) []SchedulePeriodView {
	views := make([]SchedulePeriodView, 0)
	for _, period := range periods {
		for _, day := range sharedmodels.DayList {
			if period.DaysBitmask.IsDaySet(day) {
				views = append(views, SchedulePeriodView{
					Period:   periodDTO.NewPeriodView(&period),
					Position: NewGridPosition(period.StartTime, day),
				})
			}
		}
	}
	return views
}

func timeToRow(t sharedmodels.TimeOnly, offset int) int {
	tt := time.Time(t)
	totalMinutes := tt.Hour()*60 + tt.Minute()
	return totalMinutes - offset
}

func dayToColumn(d sharedmodels.Day) int {
	columns := map[time.Weekday]int{
		time.Monday:    1,
		time.Tuesday:   2,
		time.Wednesday: 3,
		time.Thursday:  4,
		time.Friday:    5,
	}
	return columns[time.Weekday(d)]
}
