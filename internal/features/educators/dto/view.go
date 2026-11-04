package dto

import (
	"seek/internal/features/educators/models"
	periodModels "seek/internal/features/periods/models"
	studentModels "seek/internal/features/students/models"
)

type EducatorView struct {
	models.Educator
	Periods        []periodModels.Period
	PeriodStudents map[string][]studentModels.Student
	Caseload       []studentModels.Student
}

func NewView(educator *models.Educator) EducatorView {
	if educator == nil {
		return EducatorView{}
	}
	return EducatorView{Educator: *educator}
}

func NewViews(educators []models.Educator) []EducatorView {
	views := make([]EducatorView, len(educators))
	for i := range educators {
		views[i] = NewView(&educators[i])
	}
	return views
}
