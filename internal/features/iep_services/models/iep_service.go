package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type IEPService struct {
	ID              string                   `json:"id"`
	StudentID       string                   `json:"student_id" csv:"MARSS ID"`
	ServiceName     string                   `json:"service_name" csv:"Service"`
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes" csv:"Indirect minutes"`
	DirectMinutes   int                      `json:"direct_minutes" csv:"Direct minutes"`
	FrequencyCount  int                      `json:"frequency_count" csv:"Frequency count"`
	FrequencyType   string                   `json:"frequency_type" csv:"Frequency"`
	Location        string                   `json:"location"`
	StartDate       sharedmodels.DateOnly    `json:"start_date" csv:"Start date"`
	EndDate         sharedmodels.DateOnly    `json:"end_date" csv:"End date"`
	Provider        string                   `json:"provider" csv:"Provider"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewIEPService() *IEPService {
	return &IEPService{}
}
