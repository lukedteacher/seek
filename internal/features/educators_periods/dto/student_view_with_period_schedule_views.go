package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	periodDTO "seek/internal/features/periods/dto"
)

type EducatorViewWithPeriodScheduleViews struct {
	Educator educatorDTO.EducatorView
	Periods  []periodDTO.PeriodScheduleView
}
