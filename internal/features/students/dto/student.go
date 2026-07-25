package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/events"
	"seek/internal/features/students/models"
)

type StudentView struct {
	ID          string             `json:"id"`
	GivenName   string             `json:"first_name"`
	ChosenName  string             `json:"chosen_name"`
	FamilyName  string             `json:"last_name"`
	Grade       sharedmodels.Grade `json:"grade"`
	Homeroom    string             `json:"homeroom"`
	CaseManager string             `json:"case_manager"`
}

type StudentFormView struct {
	Student    StudentView                  `json:"student"`
	Validation map[string]events.Validation `json:"validation"`
}

func NewStudentViewFromModel(s models.Student) *StudentView {
	return &StudentView{
		ID:          s.ID,
		GivenName:   s.GivenName,
		ChosenName:  s.ChosenName,
		FamilyName:  s.FamilyName,
		Grade:       s.Grade,
		Homeroom:    s.Homeroom,
		CaseManager: s.CaseManager,
	}
}

func NewStudentFormViewFromModel(s models.Student) *StudentFormView {
	student := NewStudentViewFromModel(s)
	validation := events.Validate(&s)
	return &StudentFormView{
		Student:    *student,
		Validation: validation,
	}
}

func NewStudentModelFromView(v *StudentView) *models.Student {
	m := models.NewStudent()
	m.ID = v.ID
	m.GivenName = v.GivenName
	if v.ChosenName != "" {
		m.ChosenName = v.ChosenName
	}
	m.FamilyName = v.FamilyName
	m.Grade = v.Grade
	m.Homeroom = v.Homeroom
	if v.CaseManager != "" {
		m.CaseManager = v.CaseManager
	}
	return m
}
