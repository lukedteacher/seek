package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	emodels "seek/internal/features/educators/models"
	"seek/internal/features/students/models"
)

type StudentView struct {
	ID                  string                `json:"id"`
	MARSSID             string                `json:"marss_id"`
	sharedmodels.Person                       // embeds given, chosen, & family name and email & username fields
	Grade               sharedmodels.Grade    `json:"grade"`
	HomeroomID          string                `json:"homeroom_id"`
	PlanType            sharedmodels.PlanType `json:"plan_type"`
	CaseManagerView     edto.EducatorView
}

func NewStudentView(s *models.Student, e *emodels.Educator) StudentView {
	if s == nil {
		return StudentView{}
	}
	cmv := edto.EducatorView{}
	if e != nil {
		cmv = edto.NewEducatorView(e)
	}
	return StudentView{
		ID:              s.ID,
		MARSSID:         s.MARSSID,
		Person:          s.Person,
		Grade:           s.Grade,
		HomeroomID:      s.HomeroomID,
		PlanType:        s.PlanType,
		CaseManagerView: cmv,
	}
}

func NewStudentViews(students []models.Student) []StudentView {
	studentViews := make([]StudentView, len(students))
	for i, s := range students {
		studentViews[i] = StudentView{
			ID:         s.ID,
			MARSSID:    s.MARSSID,
			Person:     s.Person,
			Grade:      s.Grade,
			HomeroomID: s.HomeroomID,
			PlanType:   s.PlanType,
		}
	}
	return studentViews
}

func NewStudentModelFromView(v *StudentView) models.Student {
	if v == nil {
		return models.Student{}
	}
	return models.Student{
		ID:         v.ID,
		MARSSID:    v.MARSSID,
		Person:     v.Person,
		Grade:      v.Grade,
		HomeroomID: v.HomeroomID,
		PlanType:   v.PlanType,
	}
}
