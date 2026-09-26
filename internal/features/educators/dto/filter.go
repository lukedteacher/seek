package dto

import "seek/internal/features/_shared/sharedmodels"

type Filter struct {
	Role   map[string]bool `json:"role"`
	Search string          `json:"search"`
}

func NewFilter() Filter {
	roles := sharedmodels.EducatorRoleList
	roleFilter := make(map[string]bool, len(roles))
	for _, role := range roles {
		roleFilter[role.String()] = true
	}
	return Filter{}
}
