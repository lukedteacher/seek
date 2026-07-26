package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"time"
)

type Period struct {
	ID          string
	Title       string
	ServiceType string
	StartTime   time.Time
	EndTime     time.Time
	Duration    int64
	Days        int64
	StudentIDs  []string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
}

type PeriodSignals struct {
	ID        string                   `json:"id"`
	Title     string                   `json:"title"`
	StartTime time.Time                `json:"start_time"`
	Duration  int                      `json:"duration"`
	Days      sharedmodels.DaysSignals `json:"days"`
}

func NewPeriod() (*Period, error) {
	start, err := time.Parse("15:04", "9:30")
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("15:04", "10:00")
	if err != nil {
		return nil, err
	}
	return &Period{
		StartTime: start,
		EndTime:   end,
		Duration:  30,
	}, nil
}

func (p Period) StudentCount() int {
	return len(p.StudentIDs)
}
