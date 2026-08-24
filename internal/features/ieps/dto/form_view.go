package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	"seek/internal/features/ieps/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type IEPFormView struct {
	FormType   string
	IEPService IEPView
	Students   []studentDTO.SelectStudentOption
	Providers  []educatorDTO.EducatorSelectBoxView
}

func NewIEPFormView(
	formType string,
	model *models.IEP,
	students []studentModels.Student,
) IEPFormView {
	if model == nil {
		return IEPFormView{}
	}
	studentViews := studentDTO.NewSelectStudentOptions(students, []string{model.StudentID})
	view := NewIEPView(model)
	return IEPFormView{
		IEPService: view,
		Students:   studentViews,
	}
}
