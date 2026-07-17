package dto

import (
	"seek/internal/features/iep_services/models"
)

type IEPServiceView struct {
	IEPServiceID    string
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
}

func NewIEPServiceView(sm *models.IEPService) IEPServiceView {
	if sm == nil {
		return IEPServiceView{}
	}
	return IEPServiceView{
		IEPServiceID: sm.IEPServiceID,
		StudentID:    sm.StudentID,
		ServiceType:  sm.ServiceType,
	}
}

func NewModelFromView(v *IEPServiceView) models.IEPService {
	if v == nil {
		return models.IEPService{}
	}
	return models.IEPService{
		IEPServiceID: v.IEPServiceID,
		StudentID:    v.StudentID,
		ServiceType:  v.ServiceType,
	}
}
