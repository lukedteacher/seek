package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/iepservices/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type IEPServiceFormView struct {
	FormType   string
	IEPService IEPServiceView
	Students   []studentDTO.SelectStudentOption
	Providers  []educatorDTO.EducatorSelectBoxView
}

func NewIEPServiceFormView(
	formType string,
	model *models.IEPService,
	students []studentModels.Student,
	providers []educatorModels.Educator,
) IEPServiceFormView {
	if model == nil {
		return IEPServiceFormView{}
	}
	studentViews := studentDTO.NewSelectStudentOptions(students, []string{model.StudentID})
	providerViews := educatorDTO.NewEducatorSelectBoxViews(providers, []string{model.ProviderID})
	view := NewIEPServiceView(model)
	return IEPServiceFormView{
		IEPService: view,
		Students:   studentViews,
		Providers:  providerViews,
	}
}
