package dto

import (
	"seek/internal/features/iep_services/models"
	"seek/internal/views/dto"
)

type IEPServiceView struct {
	IEPServiceID    string
	ServiceType     string `json:"service_type"`
	IndirectMinutes int    `json:"indirect_minutes"`
	DirectMinutes   int    `json:"direct_minutes"`
	FrequencyCount  int    `json:"frequency_count"`
	FrequencyType   string `json:"frequency_type"`
	Location        string `json:"location"`
	StartDate       string `json:"start"`
	EndDate         string `json:"end"`
	Provider        string `json:"provider"`
	StudentID       string `json:"student_id"`
	StudentView     dto.StudentView
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
		ID:    v.IEPServiceID,
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
