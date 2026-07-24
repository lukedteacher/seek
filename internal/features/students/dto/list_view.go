package dto

import (
	"seek/internal/features/students/models"
)

type StudentListView struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	ChosenName  string `json:"chosen_name"`
	LastName    string `json:"last_name"`
	Grade       string `json:"grade"`
	Homeroom    string `json:"homeroom"`
	CaseManager string `json:"case_manager"`
}

func NewStudentListView(s *models.Student) *StudentListView {
	return &StudentListView{
		ID:          s.ID,
		FirstName:   s.GivenName,
		ChosenName:  s.ChosenName,
		LastName:    s.FamilyName,
		Grade:       s.GradeOrdinal(),
		Homeroom:    s.Homeroom,
		CaseManager: s.CaseManager,
	}
}
