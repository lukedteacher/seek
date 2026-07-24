package models

import (
	"seek/internal/features/_shared/sharedmodels"
)

type Period struct {
	ID         string
	Title      string
	StartTime  string
	Duration   int64
	Days       int64
	StudentIDs []string
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}

type PeriodSignals struct {
	ID        string                   `json:"id"`
	Title     string                   `json:"title"`
	StartTime string                   `json:"start_time"`
	Duration  int                      `json:"duration"`
	Days      sharedmodels.DaysSignals `json:"days"`
}

func NewPeriod() *Period {
	return &Period{
		StartTime: "9:30",
		Duration:  30,
	}
}

func (p Period) StudentCount() int {
	return len(p.StudentIDs)
}
