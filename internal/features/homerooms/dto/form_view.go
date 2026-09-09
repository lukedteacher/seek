package dto

import (
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	"seek/internal/features/homerooms/models"
	studentDTO "seek/internal/features/students/dto"
	studentModels "seek/internal/features/students/models"
)

type HomeroomFormView struct {
	FormType string
	models.Homeroom
	Educators          map[string]bool        `json:"educators"`
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
		Homeroom:           *p,
		StudentSelectView:  studentDTO.NewSelectView(studentFilter, allStudents, p.StudentIDs),
		EducatorSelectView: educatorDTO.NewSelectView(&educatorDTO.Filter{}, allEducators, p.EducatorIDs),
	}
}

func NewHomeroomModelFromFormView(
	fv HomeroomFormView,
) models.Homeroom {
	return models.Homeroom{
		ID:            fv.ID,
		Title:         fv.Title,
		GradesBitmask: fv.GradesBitmask,
		LocationID:    fv.LocationID,
		Image:         fv.Image,
		EducatorIDs:   fv.EducatorIDs,
		StudentIDs:    fv.StudentIDs,
	}
}
