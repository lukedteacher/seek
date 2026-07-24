package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
)

type EducatorView struct {
	sharedmodels.Person        // embeds given, chosen, & family name
	ID                  string `json:"id" display:"ID"`
	Email               string `json:"email" display:"email"`
	Role                string `json:"role" display:"role"`
}

func NewEducatorView(s *models.Educator) *EducatorView {
	if s == nil {
		return nil
	}
	return &EducatorView{
		ID:     s.ID,
		Person: s.Person,
		Email:  s.Email,
		Role:   s.Role,
	}
}
