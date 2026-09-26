package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"
	"time"
)

type StudentView struct {
	ID                  string                `json:"id"`
	MARSSID             string                `json:"marss_id" csv:"MARSS ID"`
	sharedmodels.Person                       // embeds given, chosen, & family name, and email fields
	Grade               sharedmodels.Grade    `json:"grade" csv:"grade"`
	HomeroomID          string                `json:"homeroom_id"`
	PlanType            sharedmodels.PlanType `json:"plan_type" csv:"plan type int"`
	CaseManagerID       string                `json:"case_manager_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Bookmarked          bool
}

func NewView(s *models.Student) StudentView {
	if s == nil {
		return StudentView{}
	}
	return StudentView{
		ID:            s.ID,
		MARSSID:       s.MARSSID,
		Person:        s.Person,
		Grade:         s.Grade,
		PlanType:      s.PlanType,
		CaseManagerID: s.CaseManagerID,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
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
		ID:            v.ID,
		MARSSID:       v.MARSSID,
		Person:        v.Person,
		Grade:         v.Grade,
		HomeroomID:    v.HomeroomID,
		PlanType:      v.PlanType,
		CaseManagerID: v.CaseManagerID,
	}
}
