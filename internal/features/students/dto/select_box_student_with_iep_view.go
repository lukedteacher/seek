package dto

import "seek/internal/features/students/models"

type SelectStudentWithIEPOption struct {
	Student    models.StudentWithIEP
	IsSelected bool
}

func NewSelectStudentWithIEPOption(
	s models.StudentWithIEP,
	isSelected bool,
) SelectStudentWithIEPOption {
	return SelectStudentWithIEPOption{
		Student:    s,
		IsSelected: isSelected,
	}
}

func NewSelectStudentWithIEPOptions(students []models.StudentWithIEP, selected []string) []SelectStudentWithIEPOption {
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	view := make([]SelectStudentWithIEPOption, len(students))
	for i, student := range students {
		view[i] = NewSelectStudentWithIEPOption(
			student,
			selectedMap[student.ID],
		)
	}
	return view
}
