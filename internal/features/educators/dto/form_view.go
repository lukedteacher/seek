package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
)

type EducatorFormView struct {
	sharedmodels.Person                                       // embeds given, chosen, & family name, username & email fields
	ID                  string                                `json:"id"`
	Role                string                                `json:"role"`
	Roles               []shareddto.EducatorRoleSelectBoxView `json:"roles"`
}

func NewEducatorFormView(e *models.Educator) EducatorFormView {
	if e == nil {
		return EducatorFormView{}
	}
	roles := shareddto.NewEducatorRoleSelectBoxViews(
		sharedmodels.EducatorRoleList,
		e.Roles,
	)
	return EducatorFormView{
		ID:     e.ID,
		Person: e.Person,
		Roles:  roles,
	}
}
