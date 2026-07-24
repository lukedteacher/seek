package dto

import (
	"seek/internal/features/iep_services/events"
	"seek/internal/features/iep_services/models"
	sb "seek/internal/features/students/blocks"
	sm "seek/internal/features/students/models"
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
