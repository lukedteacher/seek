package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	iepModels "seek/internal/features/ieps/models"
	"seek/internal/features/services/models"
	studentModels "seek/internal/features/students/models"
)

type ServiceFormView struct {
	FormType           sharedmodels.FormType
	Service            models.Service
	IEP                iepModels.IEP
	Student            studentModels.Student
	SelectedProvider   educatorModels.Educator
	ProviderSelectView educatorDTO.SelectView
}

func NewServiceFormView(
	formType sharedmodels.FormType,
	service *models.Service,
	iep *iepModels.IEP,
	student *studentModels.Student,
	selectedProvider *educatorModels.Educator,
	providers *[]educatorModels.Educator,
) ServiceFormView {
	if service == nil {
		service = models.NewService()
	}
	if iep == nil {
		iep = iepModels.NewIEP()
	}
	if student == nil {
		student = studentModels.NewStudent()
	}
	if selectedProvider == nil {
		selectedProvider = &educatorModels.Educator{}
	}
	providerSelectView := educatorDTO.SelectView{}
	if providers == nil {
		providerSelectView = educatorDTO.NewSelectView(nil, []educatorModels.Educator{}, selectedProvider.ID)
	} else {
		providerSelectView = educatorDTO.NewSelectView(nil, *providers, selectedProvider.ID)
	}
	return ServiceFormView{
		FormType:           formType,
		Service:            *service,
		IEP:                *iep,
		Student:            *student,
		ProviderSelectView: providerSelectView,
	}
}
