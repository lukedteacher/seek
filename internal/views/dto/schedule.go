package dto

type ScheduleView struct {
	ID      string       `json:"id:`
	Title   string       `json:"title"`
	Periods []PeriodView `json:"periods"`
}
