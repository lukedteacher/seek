package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/pkg/templui/components/icon"
	"time"

	"github.com/a-h/templ"
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
	StartDate       time.Time                `json:"start_date" csv:"Start date"`
	EndDate         time.Time                `json:"end_date" csv:"End date"`
	Provider        string                   `json:"provider" csv:"Provider"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewIEPService() *IEPService {
	return &IEPService{}
}

func (s IEPService) Icon() templ.Component {
	return icon.Icon(s.ServiceType.IconName())(icon.Props{Size: "16"})
}
