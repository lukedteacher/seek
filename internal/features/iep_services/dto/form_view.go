package dto

import (
	"seek/internal/features/iep_services/events"
	"seek/internal/features/iep_services/models"
	dto "seek/internal/features/students/dto"
	sm "seek/internal/features/students/models"
)

type IEPServiceFormView struct {
	IEPService IEPServiceView
	Validation map[string]events.Validation
	Students   []dto.StudentSelectBoxView
	URL        string
}

func NewIEPServiceFormView(model *models.IEPService, students []sm.Student) IEPServiceFormView {
	if model == nil {
		return IEPServiceFormView{}
	}
	studentViews := dto.NewStudentSelectBoxViews(students, []string{model.StudentID})
	view := NewIEPServiceView(model)
	return IEPServiceFormView{
		IEPService: view,
		Validation: events.Validate(model),
		Students:   studentViews,
	}
}
