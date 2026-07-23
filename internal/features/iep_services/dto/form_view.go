package dto

import (
	sm "seek/internal/domain/models"
	"seek/internal/features/iep_services/events"
	"seek/internal/features/iep_services/models"
	sb "seek/internal/features/students/blocks"
)

type IEPServiceFormView struct {
	IEPService IEPServiceView
	Validation map[string]events.Validation
	Students   []sb.StudentMultiselectView
	URL        string
}

func NewIEPServiceFormView(model *models.IEPService, students []sm.Student) IEPServiceFormView {
	if model == nil {
		return IEPServiceFormView{}
	}
	studentViews := sb.NewStudentMultiselectView(students, []string{model.StudentID})
	view := NewIEPServiceView(model)
	return IEPServiceFormView{
		IEPService: view,
		Validation: events.Validate(model),
		Students:   studentViews,
	}
}
