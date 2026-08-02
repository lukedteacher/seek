package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	"seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
)

type PeriodView struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	EndTime     sharedmodels.TimeOnly    `json:"end_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	URL         string                   `json:"url"`
	Educators   []edto.EducatorView
	Students    []sdto.StudentView
}

func NewPeriodView(p *models.Period) PeriodView {
	if p == nil {
		return PeriodView{}
	}
	return PeriodView{
		ID:          p.ID,
		Title:       p.Title,
		ServiceType: p.ServiceType,
		StartTime:   p.StartTime,
		EndTime:     p.StartTime.Add(p.Duration),
		Duration:    p.Duration,
		DaysBitmask: p.DaysBitmask,
	}
}

func NewPeriodModelFromView(pv *PeriodView) models.Period {
	if pv == nil {
		return models.Period{}
	}
	return models.Period{
		ID:          pv.ID,
		Title:       pv.Title,
		ServiceType: pv.ServiceType,
		StartTime:   pv.StartTime,
		EndTime:     pv.EndTime,
		Duration:    pv.Duration,
		DaysBitmask: pv.DaysBitmask,
	}
}
