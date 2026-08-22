package dto

import (
	"seek/internal/features/iepservices/models"
	dto "seek/internal/features/students/dto"
	sm "seek/internal/features/students/models"
)

type IEPServiceFormView struct {
	IEPService IEPServiceView
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
		Students:   studentViews,
	}
}
