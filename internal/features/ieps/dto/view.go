package dto

import (
	"seek/internal/features/ieps/models"
)

type IEPView struct {
	models.IEP
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
		MeetingDate: v.IEP.MeetingDate,
		AmendedDate: v.IEP.AmendedDate,
	}
}
