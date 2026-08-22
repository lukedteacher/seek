package dto

import "seek/internal/features/students/models"

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

func NewSelectStudentOptions(students []models.Student, selected []string) []SelectStudentOption {
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	view := make([]SelectStudentOption, len(students))
	for i, student := range students {
		view[i] = NewSelectStudentOption(
			student,
			selectedMap[student.ID],
		)
	}
	return view
}
