package dto

import "seek/internal/features/educators/models"

type EducatorSelectBoxView struct {
	Educator EducatorView
	Selected bool
}

func NewEducatorSelectBoxView(s models.Educator, selected bool) EducatorSelectBoxView {
	return EducatorSelectBoxView{
		Educator: NewEducatorView(&s),
		Selected: selected,
	}
}

func NewEducatorSelectBoxViews(educators []models.Educator, selected []string) []EducatorSelectBoxView {
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	view := make([]EducatorSelectBoxView, len(educators))
	for i, educator := range educators {
		view[i] = NewEducatorSelectBoxView(
			educator,
			selectedMap[educator.ID],
		)
	}
	return view
}
