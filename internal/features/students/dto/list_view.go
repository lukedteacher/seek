package dto

import (
	"seek/internal/features/students/models"
)

type StudentListView struct {
	ID            string `json:"id"`
	MARSSID       string `json:"marss_id"`
	GivenName     string `json:"given_name"`
	ChosenName    string `json:"chosen_name"`
	FamilyName    string `json:"family_name"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	Grade         string `json:"grade"`
	HomeroomID    string `json:"homeroom_id"`
	CaseManagerID string `json:"case_manager_id"`
}

func NewStudentListView(s *models.Student) *StudentListView {
	return &StudentListView{
		ID:            s.ID,
		GivenName:     s.GivenName,
		ChosenName:    s.ChosenName,
		FamilyName:    s.FamilyName,
		Grade:         s.Grade.Ordinal(),
		HomeroomID:    s.HomeroomID,
		CaseManagerID: s.CaseManagerID,
	}
}
