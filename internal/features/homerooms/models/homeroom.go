package models

import "seek/internal/features/_shared/sharedmodels"

type Homeroom struct {
	ID            string                     `json:"id"`
	Title         string                     `json:"title"`
	GradesBitmask sharedmodels.GradesBitmask `json:"grades_bitmask"`
	LocationID    string                     `json:"locationID"`
	Image         string                     `json:"image"`
	EducatorIDs   []string                   `json:"educator_ids"`
	StudentIDs    []string                   `json:"student_ids"`
	CreatedAt     string                     `json:"created_at,omitempty"`
	UpdatedAt     string                     `json:"updated_at,omitempty"`
}

func NewHomeroom() Homeroom {
	return Homeroom{}
}
