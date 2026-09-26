package dto

import (
	"strconv"

	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/events"
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

func (f *Filter) Options() []events.ListOption {
	var opts []events.ListOption
	if f == nil {
		return opts
	}
	if len(f.Grade) > 0 {
		grades := []int{}
		for g, ok := range f.Grade {
			if ok {
				if i, err := strconv.Atoi(g); err == nil {
					grades = append(grades, i)
				}
			}
		}
		if len(grades) > 0 {
			opts = append(opts, events.WithGradeFilter(grades))
		}
	}
	if len(f.PlanType) > 0 {
		planTypes := []int{}
		for p, ok := range f.PlanType {
			if ok {
				if i, err := strconv.Atoi(p); err == nil {
					planTypes = append(planTypes, i)
				}
			}
		}
		if len(planTypes) > 0 {
			opts = append(opts, events.WithPlanFilter(planTypes))
		}
	}
	if f.Search != "" {
		opts = append(opts, events.WithSearchFilter(f.Search))
	}
	return opts
}
