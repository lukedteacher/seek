package models

import "seek/internal/features/_shared/sharedmodels"

type Filters struct {
	Search string     `json:"search"`
	Roles  RoleFilter `json:"roles"`
}

type RoleFilter map[sharedmodels.EducatorRole]bool

func NewFilters() *Filters {
	return &Filters{
		Roles: NewRoleFilters(),
	}
}

func NewRoleFilters() RoleFilter {
	roleFilter := make(map[sharedmodels.EducatorRole]bool)
	for _, role := range sharedmodels.EducatorRoleList {
		roleFilter[role] = true
	}
	return roleFilter
}

func ModelToString(roleFilter RoleFilter) []string {
	stringFilters := make([]string, 0)
	for role, active := range roleFilter {
		if active {
			stringFilters = append(stringFilters, role.StringLong())
		}
	}
	return stringFilters
}
