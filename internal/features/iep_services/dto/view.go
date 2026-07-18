package dto

import (
	"seek/internal/features/iep_services/models"
	"seek/internal/views/dto"
)

type IEPServiceView struct {
	IEPServiceID    string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
	StudentView     dto.StudentView
}

func NewIEPServiceView(sm *models.IEPService) IEPServiceView {
	if sm == nil {
		return IEPServiceView{}
	}
	return IEPServiceView{
		IEPServiceID:    sm.IEPServiceID,
		ServiceType:     sm.ServiceType,
		IndirectMinutes: sm.IndirectMinutes,
		DirectMinutes:   sm.DirectMinutes,
		FrequencyCount:  sm.FrequencyCount,
		FrequencyType:   sm.FrequencyType,
		Location:        sm.Location,
		StartDate:       sm.StartDate,
		EndDate:         sm.EndDate,
		Provider:        sm.Provider,
	}
}

func NewModelFromView(v *IEPServiceView) models.IEPService {
	if v == nil {
		return models.IEPService{}
	}
	return models.IEPService{
		IEPServiceID:    v.IEPServiceID,
		ServiceType:     v.ServiceType,
		IndirectMinutes: v.IndirectMinutes,
		DirectMinutes:   v.DirectMinutes,
		FrequencyCount:  v.FrequencyCount,
		FrequencyType:   v.FrequencyType,
		Location:        v.Location,
		StartDate:       v.StartDate,
		EndDate:         v.EndDate,
		Provider:        v.Provider,
	}
}
