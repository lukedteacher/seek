package dto

import "seek/internal/features/students/models"

type StudentSelectBoxView struct {
	Student  StudentView
	Selected bool
}

func NewStudentSelectBoxView(s models.Student, selected bool) StudentSelectBoxView {
	return StudentSelectBoxView{
		Student:  NewStudentView(&s),
		Selected: selected,
	}
}

func NewStudentSelectBoxViews(students []models.Student, selected []string) []StudentSelectBoxView {
	selectedMap := make(map[string]bool, len(selected))
	for i := range selected {
		selectedMap[selected[i]] = true
	}
	view := make([]StudentSelectBoxView, len(students))
	for i, student := range students {
		view[i] = NewStudentSelectBoxView(
			student,
			selectedMap[student.ID],
		)
	}
	return view
}
