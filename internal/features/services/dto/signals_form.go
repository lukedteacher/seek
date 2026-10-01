package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/services/models"
)

type ServiceFormSignals struct {
	FormType       sharedmodels.FormType `json:"form_type"`
	Service        models.Service        `json:"service"`
	EducatorSelect EducatorSelectSignals `json:"educator_select"`
}

type EducatorSelectSignals struct {
	Filter educatorDTO.Filter `json:"filter"`
}

func NewServiceFormSignals(formType sharedmodels.FormType, model *models.Service) *ServiceFormSignals {
	service := models.Service{}
	if model != nil {
		service = *model
	}
	return &ServiceFormSignals{
		FormType: formType,
		Service:  service,
		EducatorSelect: EducatorSelectSignals{
			Filter: educatorDTO.NewFilter(),
		},
	}
}
