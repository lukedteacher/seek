package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/homerooms/models"
)

type HomeroomView struct {
	ID            string                     `json:"id"`
	Title         string                     `json:"title"`
	GradesBitmask sharedmodels.GradesBitmask `json:"grades_bitmask"`
	LocationID    string                     `json:"locationID"`
	Image         string                     `json:"image"`
	EducatorIDs   []string                   `json:"educator_ids"`
	StudentIDs    []string                   `json:"student_ids"`
	CreatedAt     string                     `json:"created_at,omitempty"`
	UpdatedAt     string                     `json:"updated_at,omitempty"`
}

// must initialize empty strings so signals save correctly in viewstore
func NewHomeroomView(m *models.Homeroom) HomeroomView {
	if m == nil {
		return HomeroomView{
			EducatorIDs: []string{},
			StudentIDs:  []string{},
		}
	}
	return HomeroomView{
		ID:            m.ID,
		Title:         m.Title,
		GradesBitmask: m.GradesBitmask,
		LocationID:    m.LocationID,
		Image:         m.Image,
		EducatorIDs:   m.EducatorIDs,
		StudentIDs:    m.StudentIDs,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func NewHomeroomModelFromView(v HomeroomView) models.Homeroom {
	return models.Homeroom{
		ID:            v.ID,
		Title:         v.Title,
		GradesBitmask: v.GradesBitmask,
		LocationID:    v.LocationID,
		Image:         v.Image,
		EducatorIDs:   v.EducatorIDs,
		StudentIDs:    v.StudentIDs,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}
