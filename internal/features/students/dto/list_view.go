package dto

import (
	"seek/internal/features/students/models"
)

type StudentListView struct {
	ID          string `json:"id"`
	GivenName   string `json:"given_name"`
	ChosenName  string `json:"chosen_name"`
	FamilyName  string `json:"family_name"`
	Grade       string `json:"grade"`
	Homeroom    string `json:"homeroom"`
	CaseManager string `json:"case_manager"`
}

func NewStudentListView(s *models.Student) *StudentListView {
	return &StudentListView{
		ID:          s.ID,
		GivenName:   s.GivenName,
		ChosenName:  s.ChosenName,
		FamilyName:  s.FamilyName,
		Grade:       s.Grade.Ordinal(),
		Homeroom:    s.Homeroom,
		CaseManager: s.CaseManager,
	}
}
