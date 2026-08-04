package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/dto"
	"seek/internal/features/periods/models"
)

type PersonWithScheduleView struct {
	Person    sharedmodels.Person
	Periods   []dto.PeriodScheduleView
	Color     string
	IsVisible bool
}

func NewPersonScheduleView(person sharedmodels.Person, periods []models.Period, isVisible bool, index int) PersonWithScheduleView {
	return PersonWithScheduleView{
		Person:    person,
		Periods:   dto.NewPeriodScheduleViews(periods...),
		Color:     nextColor(index),
		IsVisible: isVisible,
	}
}

var colorPalette = []string{
	"#f4433680",
	"#e91e63",
	"#9c27b0",
	"#3f51b5",
	"#2196f3",
	"#009688",
	"#4caf50",
	"#ffeb3b",
	"#ff9800",
	"#795548",
}

func nextColor(index int) string {
	return colorPalette[index%len(colorPalette)]
}
