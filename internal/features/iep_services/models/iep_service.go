package models

type IEPService struct {
	ID              string `display:"ID"`
	StudentID       string `display:"student ID" csv:"MARSS ID"`
	ServiceType     string `display:"type" csv:"Service"`
	IndirectMinutes int    `display:"indirect (min)" csv:"Indirect minutes"`
	DirectMinutes   int    `display:"direct (min)" csv:"Direct minutes"`
	FrequencyCount  int    `display:"freq count" csv:"Frequency count"`
	FrequencyType   string `display:"freq type" csv:"Frequency"`
	Location        string `display:"location"`
	StartDate       string `display:"start date" csv:"Start date"`
	EndDate         string `display:"end date" csv:"End date"`
	Provider        string `display:"provider" csv:"Provider"`
	CreatedAt       string
	UpdatedAt       string
	ArchivedAt      string
}

type IEPServiceSignals struct {
	ID              string `json:"id"`
	StudentID       string `json:"student_id"`
	ServiceType     string `json:"service_type"`
	IndirectMinutes int    `json:"indirect_minutes"`
	DirectMinutes   int    `json:"direct_minutes"`
	FrequencyCount  int    `json:"frequency_count"`
	FrequencyType   string `json:"frequency_type"`
	Location        string `json:"location"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	Provider        string `json:"provider"`
}

func NewIEPService() *IEPService {
	return &IEPService{}
}
