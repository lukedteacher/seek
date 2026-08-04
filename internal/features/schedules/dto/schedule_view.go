package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/dto"
)

type ScheduleView struct {
	Person  sharedmodels.Person
	Periods []dto.PeriodScheduleView
	Color   string
}
