package dto

import (
	sm "seek/internal/features/_shared/models"
	"seek/internal/features/educators/models"
)

type EducatorView struct {
	sm.Person        // embeds given, chosen, & family name
	ID        string `json:"id" display:"ID"`
	Email     string `json:"email" display:"email"`
	Role      string `json:"role" display:"role"`
}

func NewEducatorView(s *models.Educator) *EducatorView {
	if s == nil {
		return nil
	}
	return &EducatorView{
		ID:     s.ID,
		Person: s.Person,
		Role:   s.Role,
		Email:  s.Email,
	}
}
