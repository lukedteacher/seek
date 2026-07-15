package dto

import (
	"strconv"

	"seek/internal/domain/models"
)

type StudentView struct {
	ID          string       `json:"id"`
	FirstName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	LastName    string       `json:"last_name"`
	Grade       string       `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	Schedule    ScheduleView `json:"schedule"`
}

func NewStudentView(s *models.Student) *StudentView {
	chosenName := ""
	if s.ChosenName != nil {
		chosenName = *s.ChosenName
	}
	caseManager := ""
	if s.CaseManager != nil {
		caseManager = *s.CaseManager
	}
	return &StudentView{
		ID:          s.Id,
		FirstName:   s.FirstName,
		ChosenName:  chosenName,
		LastName:    s.LastName,
		Grade:       strconv.Itoa(int(s.Grade)),
		Homeroom:    s.Homeroom,
		CaseManager: caseManager,
	}
}
