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
	EducatorIDs string                   `json:"educator_ids"`
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

// updates the start time and recalculates the end time
// based on the current duration.
func (p *Period) UpdateStartTime(newStart sharedmodels.TimeOnly) {
	p.StartTime = newStart
	p.EndTime = sharedmodels.TimeOnly(
		time.Time(newStart).Add(time.Duration(p.Duration) * time.Minute),
	)
}

// updates the end time and recalculates the duration
// based on the current start time.
func (p *Period) UpdateEndTime(newEnd sharedmodels.TimeOnly) {
	// Calculate difference between new end and current start
	diff := time.Time(newEnd).Sub(time.Time(p.StartTime))
	if diff >= 0 {
		// Normal case: end is after start – update duration
		p.Duration = int(diff.Minutes())
		p.EndTime = newEnd
	} else {
		// End moved earlier than start – shift the whole period earlier
		// Keep duration constant, compute new start = newEnd - duration
		newStart := time.Time(newEnd).Add(-time.Duration(p.Duration) * time.Minute)
		p.StartTime = sharedmodels.TimeOnly(newStart)
		p.EndTime = newEnd
		// Duration stays the same
	}
}

// updates the duration and recalculates the end time
// based on the current start time.
func (p *Period) UpdateDuration(newDuration int) {
	if newDuration < 0 {
		newDuration = 0
	}
	p.Duration = newDuration
	p.EndTime = sharedmodels.TimeOnly(
		time.Time(p.StartTime).Add(time.Duration(newDuration) * time.Minute),
	)
}
