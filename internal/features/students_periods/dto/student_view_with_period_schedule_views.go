package dto

import (
	periodDTO "seek/internal/features/periods/dto"
	periodModels "seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type StudentWithPeriodsView struct {
	Student  studentDTO.StudentView
	Periods  []periodDTO.PeriodScheduleView
	Selected bool
}

func NewStudentWithPeriodsView(student studentModels.Student, periods []periodModels.Period, isSelected bool, index int) StudentWithPeriodsView {
	return StudentWithPeriodsView{
		Student:  studentDTO.NewView(&student),
		Periods:  periodDTO.NewPeriodScheduleViews(periods...),
		Selected: isSelected,
	}
}
