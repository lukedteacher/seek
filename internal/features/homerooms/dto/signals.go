package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
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
	Filter educatorDTO.Filter `json:"filter"`
}

type StudentSelectSignals struct {
	Filter studentDTO.Filter `json:"filter"`
}

func NewHomeroomFormSignals(formType sharedmodels.FormType, model *models.Homeroom) *HomeroomFormSignals {
	return &HomeroomFormSignals{
		FormType: formType,
		Homeroom: NewHomeroomView(model),
		EducatorSelect: EducatorSelectSignals{
			Filter: educatorDTO.NewFilter(),
		},
		StudentSelect: StudentSelectSignals{
			Filter: studentDTO.NewFilter(),
		},
	}
}
