package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type Service struct {
	ID              string                   `json:"id"`
	IEPID           string                   `json:"iep_id"`
	StudentID       string                   `json:"student_id"`
	StudentMARSSID  string                   `json:"student_marss_id" csv:"MARSS ID"`
	ServiceName     string                   `json:"service_name" csv:"Service"`
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes" csv:"Indirect minutes"`
	DirectMinutes   int                      `json:"direct_minutes" csv:"Direct minutes"`
	FrequencyCount  int                      `json:"frequency_count" csv:"Frequency count"`
	FrequencyType   string                   `json:"frequency_type" csv:"Frequency"`
	LocationID      string                   `json:"location_id" csv:"Location"`
	StartDate       sharedmodels.DateOnly    `json:"start_date" csv:"Start date"`
	EndDate         sharedmodels.DateOnly    `json:"end_date" csv:"End date"`
	Provider        string                   `json:"provider" csv:"Provider"`
	ProviderID      string                   `json:"provider_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewService() *Service {
	return &Service{}
}
