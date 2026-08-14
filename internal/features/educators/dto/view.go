package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
)

type EducatorView struct {
	sharedmodels.Person                             // embeds given, chosen, & family name
	ID                  string                      `json:"id"`
	Roles               []sharedmodels.EducatorRole `json:"roles"`
}

func NewEducatorView(e *models.Educator) EducatorView {
	if e == nil {
		return EducatorView{}
	}
	return EducatorView{
		ID:     e.ID,
		Person: e.Person,
		Roles:  e.Roles,
	}
}

func NewEducatorViews(educators []models.Educator) []EducatorView {
	views := make([]EducatorView, len(educators))
	for i := range educators {
		views[i] = NewEducatorView(&educators[i])
	}
	return views
}

func (m *EducatorView) IsCaseManager() bool {
	for _, role := range m.Roles {
		if role == sharedmodels.EducatorRoleCaseManager {
			return true
		}
	}
	return false
}
