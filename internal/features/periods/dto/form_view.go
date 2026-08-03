package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	edto "seek/internal/features/educators/dto"
	"seek/internal/features/periods/events"
	"seek/internal/features/periods/models"
	sdto "seek/internal/features/students/dto"
	smodels "seek/internal/features/students/models"
	"strings"
)

type PeriodFormView struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	EndTime     sharedmodels.TimeOnly    `json:"end_time"`
	Duration    int                      `json:"duration"`
	Days        shareddto.DaysFormView   `json:"days"`
	EducatorIDs string                   `json:"educator_ids"`
	StudentIDs  string                   `json:"student_ids"`
	URL         string                   `json:"url"`
	Validation  map[string]events.Validation
	StudentList []sdto.StudentSelectBoxView
	Educators   []edto.EducatorView
}

func NewPeriodFormView(
	p *models.Period,
	allStudents []smodels.Student,
) PeriodFormView {
	if p == nil {
		return PeriodFormView{}
	}
	ssids := strings.Split(p.StudentIDs, ",")
	return PeriodFormView{
		ID:          p.ID,
		Title:       p.Title,
		ServiceType: p.ServiceType,
		StartTime:   p.StartTime,
		EndTime:     p.EndTime,
		Duration:    p.Duration,
		Days:        shareddto.DaysBitmaskToFormView(p.DaysBitmask),
		EducatorIDs: p.EducatorIDs,
		StudentIDs:  p.StudentIDs,
		Validation:  events.Validate(p),
		StudentList: sdto.NewStudentSelectBoxViews(allStudents, ssids),
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
		EducatorIDs: v.EducatorIDs,
		StudentIDs:  v.StudentIDs,
	}
}
