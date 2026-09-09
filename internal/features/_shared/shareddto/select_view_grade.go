package shareddto

import (
	"seek/internal/features/_shared/sharedmodels"
)

type SelectViewGrade struct {
	Options []SelectOptionGrade `json:"options"`
}

type SelectOptionGrade struct {
	sharedmodels.Grade
	Selected bool
}

func NewSelectViewGrade(
	selected sharedmodels.GradesBitmask,
) SelectViewGrade {
	options := make([]SelectOptionGrade, len(sharedmodels.GradeList))
	for i, grade := range sharedmodels.GradeList {
		options[i] = NewSelectOptionGrade(
			grade,
			selected.IsGradeSet(grade),
		)
	}

	return SelectViewGrade{
		Options: options,
	}
}

func NewSelectOptionGrade(
	g sharedmodels.Grade,
	selected bool,
) SelectOptionGrade {
	return SelectOptionGrade{
		Grade:    g,
		Selected: selected,
	}
}
