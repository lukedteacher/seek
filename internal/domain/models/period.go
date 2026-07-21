package models

type Period struct {
	ID         string `json:"id" display:"ID"`
	Title      string `json:"title" display:"title"`
	StartTime  string `json:"start_time" display:"start"`
	Duration   int64  `json:"duration" display:"duration (min)"`
	Days       int64  `json:"days" display:"days"`
	StudentIDs []string
	CreatedAt  string
	UpdatedAt  string
	ArchivedAt string
}

type PeriodSignals struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int         `json:"duration"`
	Days      DaysSignals `json:"days"`
}

func NewPeriod() *Period {
	return &Period{
		StartTime: "9:30",
		Duration:  30,
	}
}
