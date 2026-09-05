package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
)

type SelectView struct {
	Options []SelectOption `json:"options"`
	Filter  Filter         `json:"filter"`
}

type SelectOption struct {
	EducatorView
	Selected bool
}

func NewSelectView(
	filter *Filter,
	educators []models.Educator,
	selected []string,
) SelectView {
	if filter == nil {
		defaultRoleFilter := make(map[string]bool, len(sharedmodels.EducatorRoleList))
		for _, role := range sharedmodels.EducatorRoleList {
			defaultRoleFilter[role.String()] = true
		}
		filter = &Filter{
			Search: "",
			Role:   defaultRoleFilter,
		}
	}
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	options := make([]SelectOption, len(educators))
	for i, educator := range educators {
		options[i] = NewSelectOption(
			educator,
			selectedMap[educator.ID],
		)
	}

	return SelectView{
		Options: options,
		Filter:  *filter,
	}
}

func NewSelectOption(
	m models.Educator,
	selected bool,
) SelectOption {
	return SelectOption{
		EducatorView: NewEducatorView(&m),
		Selected:     selected,
	}
}
