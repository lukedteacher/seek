package dto

import (
	"seek/internal/features/_shared/sharedmodels"
)

type Filter struct {
	Grade    map[string]bool `json:"grade"`
	PlanType map[string]bool `json:"plan_type"`
	Search   string          `json:"search"`
}

func NewFilter() Filter {
	studentGradeFilter := make(map[string]bool, len(sharedmodels.GradeList))
	for _, grade := range sharedmodels.GradeList {
		studentGradeFilter[grade.String()] = true
	}
	studentPlanFilter := make(map[string]bool, len(sharedmodels.PlanTypeList))
	for _, planType := range sharedmodels.PlanTypeList {
		studentPlanFilter[planType.String()] = true
	}
	return Filter{
		Grade:    studentGradeFilter,
		PlanType: studentPlanFilter,
	}
}
