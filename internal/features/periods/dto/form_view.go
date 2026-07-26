package dto

import (
	"fmt"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"
	"strings"
	"time"
)

type PeriodFormView struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType string                   `json:"service_type"`
	StartTime   time.Time                `json:"start_time"`
	EndTime     time.Time                `json:"end_time"`
	Duration    int                      `json:"duration"`
	Days        sharedmodels.DaysSignals `json:"days"`
	StudentIDs  string                   `json:"student_ids"`
}

// fails if period isn't created with default start time and duration values
func NewFormViewFromPeriod(p *models.Period) (PeriodFormView, error) {
	if p == nil {
		return PeriodFormView{}, nil
	}
	if p.StartTime.IsZero() {
		return PeriodFormView{}, fmt.Errorf("start time not initialized in period")
	}
	if p.Duration == 0 {
		println("duration not initialized in period")
	}
	days := sharedmodels.DaysBitmaskToDaysSignals(p.Days)
	return PeriodFormView{
		ID:          p.ID,
		Title:       p.Title,
		ServiceType: p.ServiceType,
		StartTime:   p.StartTime,
		EndTime:     p.StartTime.Add(time.Duration(p.Duration) * time.Minute),
		Duration:    int(p.Duration),
		Days:        days,
	}, nil
}

func NewPeriodFromFormView(v *PeriodFormView) models.Period {
	if v == nil {
		return models.Period{}
	}
	return models.Period{
		ID:          v.ID,
		Title:       v.Title,
		ServiceType: v.ServiceType,
		StartTime:   v.StartTime,
		Duration:    int64(v.Duration),
		Days:        sharedmodels.DaysSignalsToDaysBitmask(v.Days),
		StudentIDs:  strings.Split(v.StudentIDs, ","),
	}
}
