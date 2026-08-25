package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/services/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type ServiceFormView struct {
	FormType  string
	Service   ServiceView
	Students  []studentDTO.SelectStudentWithIEPOption
	Providers []educatorDTO.EducatorSelectBoxView
}

func NewServiceFormView(
	formType string,
	model *models.Service,
	students []studentModels.StudentWithIEP,
	providers []educatorModels.Educator,
) ServiceFormView {
	if model == nil {
		return ServiceFormView{}
	}
	studentViews := studentDTO.NewSelectStudentWithIEPOptions(students, []string{model.StudentID})
	providerViews := educatorDTO.NewEducatorSelectBoxViews(providers, []string{model.ProviderID})
	view := NewServiceView(model)
	return ServiceFormView{
		Service:   view,
		Students:  studentViews,
		Providers: providerViews,
	}
}
