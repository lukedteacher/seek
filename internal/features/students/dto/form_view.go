package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/students/models"
)

type StudentFormView struct {
	FormType         string                 `json:"form_type"`
	Student          StudentView            `json:"student"`
	HomeroomTeachers educatorDTO.SelectView `json:"homeroom_teachers"`
	PlanTypeOptions  []SelectPlanTypeOption `json:"plan_types"`
	CaseManagers     educatorDTO.SelectView `json:"case_managers"`
}

func NewStudentFormView(
	formType string,
	s *models.Student,
) StudentFormView {
	student := NewView(s)
	return StudentFormView{
		FormType: formType,
		Student:  student,
	}
}
