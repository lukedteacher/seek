package dto

import (
	"strconv"

	"seek/internal/domain/models"
	"seek/internal/features/students/events"
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

type StudentFormView struct {
	Student    StudentView                  `json:"student"`
	Validation map[string]events.Validation `json:"validation"`
}

func NewStudentViewFromModel(s *models.Student) *StudentView {
	chosenName := ""
	if s.ChosenName != nil {
		chosenName = *s.ChosenName
	}
	caseManager := ""
	if s.CaseManager != nil {
		caseManager = *s.CaseManager
	}
	return &StudentView{
		ID:          s.ID,
		FirstName:   s.FirstName,
		ChosenName:  chosenName,
		LastName:    s.LastName,
		Grade:       strconv.Itoa(int(s.Grade)),
		Homeroom:    s.Homeroom,
		CaseManager: caseManager,
	}
}

func NewStudentFormViewFromModel(s *models.Student) *StudentFormView {
	student := NewStudentViewFromModel(s)
	validation := events.Validate(s)
	return &StudentFormView{
		Student:    *student,
		Validation: validation,
	}
}

func NewStudentModelFromView(v *StudentView) *models.Student {
	m := models.NewStudent()
	m.ID = v.ID
	m.FirstName = v.FirstName
	if v.ChosenName != "" {
		m.ChosenName = &v.ChosenName
	}
	m.LastName = v.LastName
	grade := -1
	if v.Grade != "select a grade" {
		grade, _ = strconv.Atoi(v.Grade)
	}
	m.Grade = int64(grade)
	m.Homeroom = v.Homeroom
	if v.CaseManager != "" {
		m.CaseManager = &v.CaseManager
	}
	return m
}
