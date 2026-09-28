package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
)

type PeriodFormSignals struct {
	FormType       sharedmodels.FormType `json:"form_type"`
	Period         PeriodView            `json:"period"`
	EducatorSelect EducatorSelectSignals `json:"educator_select"`
	StudentSelect  StudentSelectSignals  `json:"student_select"`
}

type EducatorSelectSignals struct {
	Filter educatorDTO.Filter `json:"filter"`
}

type StudentSelectSignals struct {
	Filter studentDTO.Filter `json:"filter"`
}

func NewPeriodFormSignals(formType sharedmodels.FormType, model *models.Period) *PeriodFormSignals {
	return &PeriodFormSignals{
		FormType: formType,
		Period:   NewPeriodView(model),
		EducatorSelect: EducatorSelectSignals{
			Filter: educatorDTO.NewFilter(),
		},
		StudentSelect: StudentSelectSignals{
			Filter: studentDTO.NewFilter(),
		},
	}
}
