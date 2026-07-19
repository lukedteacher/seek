package dto

import (
	"seek/internal/domain/models"
)

type StudentListView struct {
	ID          string       `json:"id"`
	FirstName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	LastName    string       `json:"last_name"`
	Grade       string       `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
}

func NewStudentListView(s *models.Student) *StudentListView {
	chosenName := ""
	if s.ChosenName != nil {
		chosenName = *s.ChosenName
	}
	caseManager := ""
	if s.CaseManager != nil {
		caseManager = *s.CaseManager
	}
	return &StudentListView{
		ID:          s.ID,
		FirstName:   s.FirstName,
		ChosenName:  chosenName,
		LastName:    s.LastName,
		Grade:       s.GradeOrdinal(),
		Homeroom:    s.Homeroom,
		CaseManager: caseManager,
	}
}
