package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
)

type EducatorView struct {
	sharedmodels.Person        // embeds given, chosen, & family name
	ID                  string `json:"id" display:"ID"`
	Role                string `json:"role" display:"role"`
}

func NewEducatorView(e *models.Educator) EducatorView {
	if e == nil {
		return EducatorView{}
	}
	return EducatorView{
		ID:     e.ID,
		Person: e.Person,
		Role:   e.Role,
	}
}

func NewEducatorViews(educators []models.Educator) []EducatorView {
	views := make([]EducatorView, len(educators))
	for i := range educators {
		views[i] = NewEducatorView(&educators[i])
	}
	return views
}
