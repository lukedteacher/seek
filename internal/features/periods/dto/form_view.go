package dto

import (
	"seek/internal/features/_shared/shareddto"
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type PeriodFormView struct {
	FormType string
	models.Period
	Days           shareddto.DaysFormView `json:"days"`
	StudentSelect  studentDTO.SelectView  `json:"student_select"`
	EducatorSelect educatorDTO.SelectView `json:"educator_select"`
}

func NewPeriodFormView(
	p *models.Period,
	allStudents []studentModels.Student,
	studentFilter *studentDTO.Filter,
	allEducators []educatorModels.Educator,
) PeriodFormView {
	if p == nil {
		return PeriodFormView{}
	}
	return PeriodFormView{
		Period:         *p,
		Days:           shareddto.DaysBitmaskToFormView(p.DaysBitmask),
		StudentSelect:  studentDTO.NewSelectView(studentFilter, allStudents, p.StudentIDs),
		EducatorSelect: educatorDTO.NewSelectView(&educatorDTO.Filter{}, allEducators, p.EducatorIDs),
	}
}

func NewModelFromFormView(fv PeriodFormView) models.Period {
	return models.Period{
		ID:          fv.Period.ID,
		Title:       fv.Period.Title,
		ServiceType: fv.Period.ServiceType,
		StartTime:   fv.Period.StartTime,
		EndTime:     fv.Period.EndTime,
		Duration:    fv.Period.Duration,
		DaysBitmask: fv.Days.ToBitmask(),
		EducatorIDs: fv.Period.EducatorIDs,
		StudentIDs:  fv.Period.StudentIDs,
	}
}
