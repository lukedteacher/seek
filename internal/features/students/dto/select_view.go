package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"
)

type SelectView struct {
	Options []SelectOption `json:"options"`
	Filter  Filter         `json:"filter"`
}

type SelectOption struct {
	StudentView
	Selected bool
}

func NewSelectView(
	filter *Filter,
	students []models.Student,
	selected []string,
) SelectView {
	if filter == nil {
		defaultGradeFilter := make(map[string]bool, 9)
		for _, grade := range sharedmodels.GradeList {
			defaultGradeFilter[grade.String()] = true
		}
		defaultPlanTypeFilter := make(map[string]bool, 4)
		for _, planType := range sharedmodels.PlanTypeList {
			defaultPlanTypeFilter[planType.String()] = true
		}
		filter = &Filter{
			Grade:    defaultGradeFilter,
			PlanType: defaultPlanTypeFilter,
			Search:   "",
		}
	}
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	options := make([]SelectOption, len(students))
	for i, student := range students {
		options[i] = NewSelectOption(
			student,
			selectedMap[student.ID],
		)
	}

	return SelectView{
		Options: options,
		Filter:  *filter,
	}
}

func NewSelectOption(
	s models.Student,
	selected bool,
) SelectOption {
	return SelectOption{
		StudentView: NewStudentView(&s, nil),
		Selected:    selected,
	}
}
