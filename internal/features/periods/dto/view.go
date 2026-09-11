package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
)

type PeriodView struct {
	models.Period
	Educators []educatorDTO.EducatorView
	Students  []studentDTO.StudentView
}

func NewPeriodView(p *models.Period) PeriodView {
	if p == nil {
		return PeriodView{}
	}
	return PeriodView{
		Period: *p,
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
