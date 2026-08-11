package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"
)

type StudentView struct {
	ID                  string             `json:"id"`
	MARSSID             string             `json:"marss_id"`
	sharedmodels.Person                    // embeds given, chosen, & family name and email & username fields
	Grade               sharedmodels.Grade `json:"grade"`
	Homeroom            string             `json:"homeroom"`
	CaseManager         string             `json:"case_manager"`
}

func NewStudentView(s *models.Student) StudentView {
	if s == nil {
		return StudentView{}
	}
	return StudentView{
		ID:          s.ID,
		Person:      s.Person,
		Grade:       s.Grade,
		Homeroom:    s.Homeroom,
		CaseManager: s.CaseManager,
	}
}

func NewStudentModelFromView(v *StudentView) models.Student {
	if v == nil {
		return models.Student{}
	}
	return models.Student{
		ID:          v.ID,
		Person:      v.Person,
		Grade:       v.Grade,
		Homeroom:    v.Homeroom,
		CaseManager: v.CaseManager,
	}
}
