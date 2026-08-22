package dto

import (
	"seek/internal/features/_shared/sharedmodels"
)

type SelectPlanTypeOption struct {
	PlanType   sharedmodels.PlanType
	IsSelected bool
}

func NewSelectPlanTypeOption(
	m sharedmodels.PlanType,
	isSelected bool,
) SelectPlanTypeOption {
	return SelectPlanTypeOption{
		PlanType:   m,
		IsSelected: isSelected,
	}
}

func NewSelectPlanTypeOptions(
	planTypes []sharedmodels.PlanType,
	selected sharedmodels.PlanType,
) []SelectPlanTypeOption {
	view := make([]SelectPlanTypeOption, len(planTypes))
	for i := range planTypes {
		view[i] = NewSelectPlanTypeOption(
			planTypes[i],
			planTypes[i] == selected,
		)
	}
	return view
}
