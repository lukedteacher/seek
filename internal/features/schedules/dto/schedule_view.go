package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/dto"
	"seek/internal/features/periods/models"
)

type PersonWithScheduleView struct {
	ID        string // TODO fix this
	Person    sharedmodels.Person
	Periods   []dto.PeriodScheduleView
	Color     string
	IsVisible bool
}

func NewPersonScheduleView(id string, person sharedmodels.Person, periods []models.Period, isVisible bool, index int) PersonWithScheduleView {
	return PersonWithScheduleView{
		ID:        id,
		Person:    person,
		Periods:   dto.NewPeriodScheduleViews(periods...),
		Color:     nextColor(index),
		IsVisible: isVisible,
	}
}

var colorPalette = []string{
	"default",
	"red",
	"orange",
	"yellow",
	"green",
	"aqua",
	"blue",
	"purple",
}

func nextColor(index int) string {
	return colorPalette[index%len(colorPalette)]
}
