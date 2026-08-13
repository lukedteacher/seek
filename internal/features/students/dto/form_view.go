package dto

import (
	"seek/internal/features/educators/dto"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
)

type StudentFormView struct {
	Student          StudentView                  `json:"student"`
	HomeroomTeachers []dto.EducatorSelectBoxView  `json:"homeroom_teachers"`
	CaseManagers     []dto.EducatorSelectBoxView  `json:"case_managers"`
	Validation       map[string]events.Validation `json:"validation"`
}

func NewStudentFormView(s *models.Student) StudentFormView {
	student := NewStudentView(s)
	validation := events.Validate(s)
	return StudentFormView{
		Student:    student,
		Validation: validation,
	}
}
