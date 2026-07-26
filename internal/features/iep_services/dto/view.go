package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/iep_services/models"
	sdto "seek/internal/features/students/dto"
)

type IEPServiceView struct {
	IEPServiceID    string
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes"`
	DirectMinutes   int                      `json:"direct_minutes"`
	FrequencyCount  int                      `json:"frequency_count"`
	FrequencyType   string                   `json:"frequency_type"`
	Location        string                   `json:"location"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Provider        string                   `json:"provider"`
	StudentID       string                   `json:"student_id"`
	StudentView     sdto.StudentView
}

func NewIEPServiceView(sm *models.IEPService) IEPServiceView {
	if sm == nil {
		return IEPServiceView{}
	}
	return IEPServiceView{
		IEPServiceID:    sm.ID,
		ServiceType:     sm.ServiceType,
		IndirectMinutes: sm.IndirectMinutes,
		DirectMinutes:   sm.DirectMinutes,
		FrequencyCount:  sm.FrequencyCount,
		FrequencyType:   sm.FrequencyType,
		Location:        sm.Location,
		StartDate:       sm.StartDate,
		EndDate:         sm.EndDate,
		Provider:        sm.Provider,
		StudentID:       sm.StudentID,
	}
}

func NewModelFromView(v *IEPServiceView) models.IEPService {
	if v == nil {
		return models.IEPService{}
	}
	return models.IEPService{
		ID:              v.IEPServiceID,
		ServiceType:     v.ServiceType,
		IndirectMinutes: v.IndirectMinutes,
		DirectMinutes:   v.DirectMinutes,
		FrequencyCount:  v.FrequencyCount,
		FrequencyType:   v.FrequencyType,
		Location:        v.Location,
		StartDate:       v.StartDate,
		EndDate:         v.EndDate,
		Provider:        v.Provider,
		StudentID:       v.StudentID,
	}
}
