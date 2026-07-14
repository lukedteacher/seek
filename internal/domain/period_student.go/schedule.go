package models

type Schedule struct {
	Id        string
	Title     string
	TeacherId string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}

type ScheduleSignals struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	TeacherID string   `json:"teacher_id"`
	PeriodIDs []string `json:"period_ids"`
}
