package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/dto"
)

type PeriodWithStudentsView struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	EndTime     sharedmodels.TimeOnly    `json:"end_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	Students    []dto.StudentView
}
