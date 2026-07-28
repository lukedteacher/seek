package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"time"
)

type Period struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	EndTime     sharedmodels.TimeOnly    `json:"end_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	StudentIDs  string                   `json:"student_ids"`
	CreatedAt   string
	UpdatedAt   string
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
		StartTime: sharedmodels.TimeOnly(start),
		EndTime:   sharedmodels.TimeOnly(end),
		Duration:  30,
	}, nil
}
