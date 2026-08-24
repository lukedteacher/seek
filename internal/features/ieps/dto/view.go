package dto

import (
	"seek/internal/features/ieps/models"
	sdto "seek/internal/features/students/dto"
)

type IEPView struct {
	models.IEP
	StudentView sdto.StudentView
}

func NewIEPView(m *models.IEP) IEPView {
	if m == nil {
		return IEPView{}
	}
	return IEPView{
		IEP: *m,
	}
}

func NewModelFromView(v *IEPView) models.IEP {
	if v == nil {
		return models.IEP{}
	}
	return models.IEP{
		ID:          v.IEP.ID,
		StudentID:   v.IEP.StudentID,
		StartDate:   v.IEP.StartDate,
		EndDate:     v.IEP.EndDate,
		AmendedDate: v.IEP.AmendedDate,
	}
}
