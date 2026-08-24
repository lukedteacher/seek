package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type IEP struct {
	ID          string                `json:"id"`
	StudentID   string                `json:"student_id"`
	StartDate   sharedmodels.DateOnly `json:"start_date" csv:"Start date"`
	EndDate     sharedmodels.DateOnly `json:"end_date" csv:"End date"`
	AmendedDate sharedmodels.DateOnly `json:"amended_date" csv:"Amended date"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewIEP() *IEP {
	return &IEP{}
}
