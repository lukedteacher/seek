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

func NewEducatorView(s *models.Educator) *EducatorView {
	if s == nil {
		return nil
	}
	return &EducatorView{
		ID:     s.ID,
		Person: s.Person,
		Role:   s.Role,
	}
}
