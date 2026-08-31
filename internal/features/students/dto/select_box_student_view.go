package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"
)

type SelectStudentOptions struct {
	Options []SelectStudentOption `json:"options"`
	Filter  StudentFilter         `json:"filter"`
}

type SelectStudentOption struct {
	Student    StudentView
	IsSelected bool
}

func NewSelectStudentOption(
	s models.Student,
	isSelected bool,
) SelectStudentOption {
	return SelectStudentOption{
		Student:    NewStudentView(&s, nil),
		IsSelected: isSelected,
	}
}

func NewSelectStudentOptions(filter *StudentFilter, students []models.Student, selected []string) SelectStudentOptions {
	if filter == nil {
		defaultGradeFilter := make(map[string]bool, 9)
		for _, grade := range sharedmodels.GradeList {
			defaultGradeFilter[grade.String()] = true
		}
		defaultPlanTypeFilter := make(map[string]bool, 4)
		for _, planType := range sharedmodels.PlanTypeList {
			defaultPlanTypeFilter[planType.String()] = true
		}
		filter = &StudentFilter{
			Grade:    defaultGradeFilter,
			PlanType: defaultPlanTypeFilter,
			Search:   "",
		}
	}
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	options := make([]SelectStudentOption, len(students))
	for i, student := range students {
		options[i] = NewSelectStudentOption(
			student,
			selectedMap[student.ID],
		)
	}

	return SelectStudentOptions{
		Options: options,
		Filter:  *filter,
	}
}
