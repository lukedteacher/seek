package dto

import (
	"seek/internal/features/iep_services/events"
	"seek/internal/features/iep_services/models"
)

type IEPServiceFormView struct {
	IEPService IEPServiceView
	Validation map[string]events.Validation
}

func NewStudentMinutesFormView(sm *models.IEPService) IEPServiceFormView {
	if sm == nil {
		return IEPServiceFormView{}
	}
	view := IEPServiceView{
		IEPServiceID: sm.IEPServiceID,
		StudentID:    sm.StudentID,
		ServiceType:  sm.ServiceType,
	}
	return IEPServiceFormView{
		IEPService: view,
		Validation: events.Validate(sm),
	}
}
