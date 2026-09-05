package dto

type Filter struct {
	Role   map[string]bool `json:"role"`
	Search string          `json:"search"`
}
