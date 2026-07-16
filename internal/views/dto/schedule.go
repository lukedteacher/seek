package dto

type ScheduleView struct {
	ID      string       `json:"id"`
	Title   string       `json:"title"`
	Teacher TeacherView  `json:"teacher"`
	Periods []PeriodView `json:"periods"`
}
