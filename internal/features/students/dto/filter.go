package dto

import (
	"seek/internal/features/students/events"
	"strconv"
)

type StudentFilter struct {
	Grade    map[string]bool `json:"grade"`
	PlanType map[string]bool `json:"plan_type"`
	Search   string          `json:"search"`
}

func (f *StudentFilter) Options() []events.ListOption {
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
