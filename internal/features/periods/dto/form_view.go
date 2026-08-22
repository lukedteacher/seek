package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	emodels "seek/internal/features/educators/models"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
	smodels "seek/internal/features/students/models"
	"strings"
)

type PeriodFormView struct {
	FormType        string
	ID              string                   `json:"id"`
	Title           string                   `json:"title"`
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	StartTime       sharedmodels.TimeOnly    `json:"start_time"`
	EndTime         sharedmodels.TimeOnly    `json:"end_time"`
	Duration        int                      `json:"duration"`
	Days            shareddto.DaysFormView   `json:"days"`
	EducatorIDs     string                   `json:"educator_ids"`
	StudentIDs      string                   `json:"student_ids"`
	Validation      map[string]events.Validation
	StudentOptions  []sdto.SelectStudentOption
	EducatorOptions []edto.EducatorSelectBoxView
}

func NewPeriodFormView(
	p *models.Period,
	allStudents []smodels.Student,
	allEducators []emodels.Educator,
) PeriodFormView {
	if p == nil {
		return PeriodFormView{}
	}
	return PeriodFormView{
		ID:              p.ID,
		Title:           p.Title,
		ServiceType:     p.ServiceType,
		StartTime:       p.StartTime,
		EndTime:         p.EndTime,
		Duration:        p.Duration,
		Days:            shareddto.DaysBitmaskToFormView(p.DaysBitmask),
		EducatorIDs:     strings.Join(p.EducatorIDs, ","),
		StudentIDs:      strings.Join(p.StudentIDs, ","),
		Validation:      events.Validate(p),
		StudentOptions:  sdto.NewSelectStudentOptions(allStudents, p.StudentIDs),
		EducatorOptions: edto.NewEducatorSelectBoxViews(allEducators, p.EducatorIDs),
	}
}

func (v PeriodFormView) ToPeriod() models.Period {
	return models.Period{
		ID:          v.ID,
		Title:       v.Title,
		ServiceType: v.ServiceType,
		StartTime:   v.StartTime,
		EndTime:     v.EndTime,
		Duration:    v.Duration,
		DaysBitmask: v.Days.ToBitmask(),
		EducatorIDs: strings.Split(v.EducatorIDs, ","),
		StudentIDs:  strings.Split(v.StudentIDs, ","),
	}
}
