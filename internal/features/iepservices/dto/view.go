package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/iepservices/models"
	sdto "seek/internal/features/students/dto"
)

type IEPServiceView struct {
	ID              string
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes"`
	DirectMinutes   int                      `json:"direct_minutes"`
	FrequencyCount  int                      `json:"frequency_count"`
	FrequencyType   string                   `json:"frequency_type"`
	Location        string                   `json:"location"`
	StartDate       sharedmodels.DateOnly    `json:"start_date"`
	EndDate         sharedmodels.DateOnly    `json:"end_date"`
	Provider        string                   `json:"provider"`
	StudentID       string                   `json:"student_id"`
	URL             string                   `json:"url"`
	StudentView     sdto.StudentView
}

func NewIEPServiceView(sm *models.IEPService) IEPServiceView {
	if sm == nil {
		return IEPServiceView{}
	}
	return IEPServiceView{
		ID:              sm.ID,
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
		ID:              v.ID,
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
