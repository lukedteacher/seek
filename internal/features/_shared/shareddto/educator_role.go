package shareddto

import "seek/internal/features/_shared/sharedmodels"

type EducatorRoleSelectBoxView struct {
	String     string
	IsSelected bool
}

func NewEducatorRoleSelectBoxViews(
	roles []sharedmodels.EducatorRole,
	selected []sharedmodels.EducatorRole,
) []EducatorRoleSelectBoxView {
	selectedMap := make(map[sharedmodels.EducatorRole]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	view := make([]EducatorRoleSelectBoxView, len(roles))
	for i, role := range roles {
		view[i] = EducatorRoleSelectBoxView{
			String:     role.String(),
			IsSelected: selectedMap[role],
		}
	}
	return view
}
