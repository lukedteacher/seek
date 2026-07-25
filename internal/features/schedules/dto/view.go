package dto

import (
	pdto "seek/internal/features/periods/dto"
	tdto "seek/internal/features/teachers/dto"
)

type ScheduleView struct {
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Teacher tdto.TeacherView  `json:"teacher"`
	Periods []pdto.PeriodView `json:"periods"`
}
