package dto

type SelectStudentWithIEPOption struct {
	Student  StudentWithIEP
	Selected bool
}

func NewSelectStudentWithIEPOption(
	s StudentWithIEP,
	isSelected bool,
) SelectStudentWithIEPOption {
	return SelectStudentWithIEPOption{
		Student:  s,
		Selected: isSelected,
	}
}

func NewSelectStudentWithIEPOptions(students []StudentWithIEP, selected []string) []SelectStudentWithIEPOption {
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
