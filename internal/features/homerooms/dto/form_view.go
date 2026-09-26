package dto

import (
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/homerooms/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type HomeroomFormView struct {
	FormType           sharedmodels.FormType
	Homeroom           HomeroomView
	StudentSelectView  studentDTO.SelectView  `json:"student_select"`
	EducatorSelectView educatorDTO.SelectView `json:"educator_select"`
}

func NewHomeroomFormView(
	p *models.Homeroom,
	allStudents []studentModels.Student,
	studentFilter *studentDTO.Filter,
	allEducators []educatorModels.Educator,
) HomeroomFormView {
	if p == nil {
		return HomeroomFormView{}
	}
	return HomeroomFormView{
		Homeroom:           NewHomeroomView(p),
		StudentSelectView:  studentDTO.NewSelectView(studentFilter, allStudents, p.StudentIDs),
		EducatorSelectView: educatorDTO.NewSelectView(&educatorDTO.Filter{}, allEducators, p.EducatorIDs),
	}
}

func NewHomeroomModelFromFormView(
	fv HomeroomFormView,
) models.Homeroom {
	return models.Homeroom{
		ID:            fv.Homeroom.ID,
		Title:         fv.Homeroom.Title,
		GradesBitmask: fv.Homeroom.GradesBitmask,
		LocationID:    fv.Homeroom.LocationID,
		Image:         fv.Homeroom.Image,
		EducatorIDs:   fv.Homeroom.EducatorIDs,
		StudentIDs:    fv.Homeroom.StudentIDs,
	}
}
