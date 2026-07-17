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

func NewIEPServicesView(sm *models.IEPService) IEPServiceView {
	if sm == nil {
		return IEPServiceView{}
	}
	return IEPServiceView{
		IEPServiceID: sm.IEPServiceID,
		StudentID:    sm.StudentID,
		ServiceType:  sm.ServiceType,
	}
}
