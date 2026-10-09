package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/homerooms/models"
	studentDTO "seek/internal/features/students/dto"
)

type HomeroomFormSignals struct {
	FormType       sharedmodels.FormType `json:"form_type"`
	Homeroom       HomeroomView          `json:"homeroom"`
	EducatorSelect EducatorSelectSignals `json:"educator_select"`
	StudentSelect  StudentSelectSignals  `json:"student_select"`
}

type EducatorSelectSignals struct {
	Filters educatorModels.Filters `json:"filter"`
}

type StudentSelectSignals struct {
	Filter studentDTO.Filter `json:"filter"`
}

func NewHomeroomFormSignals(formType sharedmodels.FormType, model *models.Homeroom) *HomeroomFormSignals {
	return &HomeroomFormSignals{
		FormType: formType,
		Homeroom: NewHomeroomView(model),
		EducatorSelect: EducatorSelectSignals{
			Filters: *educatorModels.NewFilters(),
		},
		StudentSelect: StudentSelectSignals{
			Filter: studentDTO.NewFilter(),
		},
	}
}
