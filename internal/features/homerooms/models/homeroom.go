package models

import "seek/internal/features/_shared/sharedmodels"

type Homeroom struct {
	ID            string                     `json:"id"`
	Title         string                     `json:"title,omitempty"`
	GradesBitmask sharedmodels.GradesBitmask `json:"grades_bitmask,omitempty"`
	LocationID    string                     `json:"locationID,omitempty"`
	Image         string                     `json:"image,omitempty"`
	EducatorIDs   []string                   `json:"educator_ids,omitempty"`
	StudentIDs    []string                   `json:"student_ids,omitempty"`
	CreatedAt     string                     `json:"created_at,omitempty"`
	UpdatedAt     string                     `json:"updated_at,omitempty"`
}

// must initialize slices to avoid null values in JSON
func NewHomeroom() Homeroom {
	return Homeroom{
		EducatorIDs: []string{},
		StudentIDs:  []string{},
	}
}
