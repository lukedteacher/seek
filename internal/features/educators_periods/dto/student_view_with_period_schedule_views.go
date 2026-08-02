package dto

import (
	sdto "seek/internal/features/educators/dto"
	pdto "seek/internal/features/periods/dto"
)

type EducatorViewWithPeriodScheduleViews struct {
	Educator sdto.EducatorView
	Periods  []pdto.PeriodScheduleView
}
