package dto

import (
	pdto "seek/internal/features/periods/dto"
	sdto "seek/internal/features/students/dto"
)

type StudentViewWithPeriodScheduleViews struct {
	Student sdto.StudentView
	Periods []pdto.PeriodScheduleView
}
