package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/homerooms/models"
	studentDTO "seek/internal/features/students/dto"
)

type HomeroomView struct {
	models.Homeroom
	Educators []educatorDTO.EducatorView
	Students  []studentDTO.StudentView
}

func NewHomeroomView(m *models.Homeroom) HomeroomView {
	if m == nil {
		return HomeroomView{}
	}
	return HomeroomView{
		Homeroom: *m,
	}
}

func NewHomeroomModelFromView(v HomeroomView) models.Homeroom {
	return models.Homeroom{
		ID:         v.ID,
		Title:      v.Title,
		LocationID: v.LocationID,
	}
}
