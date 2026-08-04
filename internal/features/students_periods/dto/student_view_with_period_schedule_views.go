package dto

import (
	pdto "seek/internal/features/periods/dto"
	pmodels "seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
	smodels "seek/internal/features/students/models"
)

type StudentWithPeriodsView struct {
	Student    sdto.StudentView
	Periods    []pdto.PeriodScheduleView
	IsSelected bool
}

func NewStudentWithPeriodsView(student smodels.Student, periods []pmodels.Period, isSelected bool, index int) StudentWithPeriodsView {
	return StudentWithPeriodsView{
		Student:    sdto.NewStudentView(&student),
		Periods:    pdto.NewPeriodScheduleViews(periods...),
		IsSelected: isSelected,
	}
}
