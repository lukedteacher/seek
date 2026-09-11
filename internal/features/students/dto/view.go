package dto

import (
	"seek/internal/features/students/models"
)

type StudentView struct {
	models.Student
	Bookmarked bool
}

func NewView(s *models.Student) StudentView {
	if s == nil {
		return StudentView{}
	}
	return StudentView{
		Student: *s,
	}
}

func NewViews(students []models.Student) []StudentView {
	studentViews := make([]StudentView, len(students))
	for i, student := range students {
		studentViews[i] = NewView(&student)
	}
	return studentViews
}

func NewModelFromView(v StudentView) models.Student {
	return models.Student{
		ID:         v.ID,
		MARSSID:    v.MARSSID,
		Person:     v.Person,
		Grade:      v.Grade,
		HomeroomID: v.HomeroomID,
		PlanType:   v.PlanType,
	}
}
