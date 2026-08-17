package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"time"
)

type CSVIEPService struct {
	ID              string                   `json:"id"`
	StudentID       string                   `json:"student_id"`
	StudentMARSSID  string                   `json:"student_marss_id" csv:"MARSS ID"`
	ServiceName     string                   `json:"service_name" csv:"Service"`
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes" csv:"Indirect minutes"`
	DirectMinutes   int                      `json:"direct_minutes" csv:"Direct minutes"`
	FrequencyCount  int                      `json:"frequency_count" csv:"Frequency count"`
	FrequencyType   string                   `json:"frequency_type" csv:"Frequency"`
	Location        string                   `json:"location"`
	StartDate       CSVTime                  `json:"start_date" csv:"Start date"`
	EndDate         CSVTime                  `json:"end_date" csv:"End date"`
	Provider        string                   `json:"provider" csv:"Provider"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CSVTime time.Time

func (ct *CSVTime) UnmarshalCSV(csv string) error {
	t, err := time.Parse("01/02/2006 03:04 PM", csv)
	if err != nil {
		return err
	}
	*ct = CSVTime(t)
	return nil
}

func NewCSVIEPService() *IEPService {
	return &IEPService{}
}

func (c CSVIEPService) ToIEPService() *IEPService {
	return &IEPService{
		ID:              c.ID,
		StudentID:       c.StudentID,
		StudentMARSSID:  c.StudentMARSSID,
		ServiceName:     c.ServiceName,
		ServiceType:     c.ServiceType,
		IndirectMinutes: c.IndirectMinutes,
		DirectMinutes:   c.DirectMinutes,
		FrequencyCount:  c.FrequencyCount,
		FrequencyType:   c.FrequencyType,
		Location:        c.Location,
		StartDate:       sharedmodels.DateOnly(c.StartDate),
		EndDate:         sharedmodels.DateOnly(c.EndDate),
		Provider:        c.Provider,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
