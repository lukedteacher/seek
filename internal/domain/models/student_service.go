package models

type StudentService struct {
	ServiceID       string
	StudentID       string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
	CreatedAt       string
	UpdatedAt       string
}

type StudentServiceSignals struct {
	ServiceID       string `json:"service_id"`
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

func NewStudentService() *StudentService {
	return &StudentService{}
}
