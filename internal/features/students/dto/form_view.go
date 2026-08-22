package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
)

type StudentFormView struct {
	FormType         string                              `json:"form_type"`
	Student          StudentView                         `json:"student"`
	HomeroomTeachers []educatorDTO.EducatorSelectBoxView `json:"homeroom_teachers"`
	PlanTypeOptions  []SelectPlanTypeOption              `json:"plan_types"`
	CaseManagers     []educatorDTO.EducatorSelectBoxView `json:"case_managers"`
	Validation       map[string]events.Validation        `json:"validation"`
}

func NewStudentFormView(
	formType string,
	s *models.Student,
) StudentFormView {
	student := NewStudentView(s, nil)
	validation := events.Validate(s)
	return StudentFormView{
		FormType:   formType,
		Student:    student,
		Validation: validation,
	}
}
