package models

type Period struct {
	ID        string
	Title     string
	StartTime string
	Duration  int64
	Days      int64
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}

type PeriodSignals struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int         `json:"duration"`
	Days      DaysSignals `json:"days"`
}

func NewPeriod() Period {
	return Period{}
}