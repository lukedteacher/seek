package dto

import "seek/internal/features/_shared/sharedmodels"

type Filter struct {
	Search string   `json:"search"`
	Role   []string `json:"role"`
}

// returns a filter that contains all roles from the role list
func NewFilter() Filter {
	roles := make([]string, len(sharedmodels.EducatorRoleList))
	for i, r := range sharedmodels.EducatorRoleList {
		roles[i] = r.StringLong()
	}
	return Filter{Role: roles}
}
