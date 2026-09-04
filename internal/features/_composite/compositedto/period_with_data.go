package compositedto

import (
	"context"
	"io"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	educatorDTO "seek/internal/features/educators/dto"
	periodModels "seek/internal/features/periods/models"
	studentDTO "seek/internal/features/students/dto"
	"seek/internal/ui/core/coreblocks"
	"strconv"

	"github.com/a-h/templ"
)

type PeriodWithData struct {
	periodModels.Period
	Educators []educatorDTO.EducatorView
	Students  []studentDTO.StudentView
}

var PeriodWithDataTableConfig = shareddto.TableConfig[PeriodWithData]{
	Name: "periods",
	Sort: shareddto.TableSort{
		Column:    "start_time",
		Direction: "ASC",
	},
	Columns:         PeriodWithDataColumns,
	ValueExtractor:  periodValueExtractor,
	TargetExtractor: periodTargetExtractor,
}

var PeriodWithDataColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "ServiceType", Display: "type", Renderer: "badge", Alignment: "center"},
	{Field: "StartTime", Display: "start", Group: "time", Alignment: "center"},
	{Field: "EndTime", Display: "end", Group: "time", Alignment: "center"},
	{Field: "Duration", Display: "duration", Alignment: "center"},
	{Field: "Days", Display: "days", Alignment: "center"},
	{Field: "Educators", Display: "educators", RenderFunc: educatorsRenderer, Signal: "educators"},
	{Field: "Students", Display: "students", RenderFunc: studentsRenderer, Signal: "students"},
}

func NewPeriodWithDataTableView(periodsWithData []PeriodWithData) shareddto.TableView {
	return shareddto.NewTableView(periodsWithData, PeriodWithDataTableConfig)
}

func periodValueExtractor(m *PeriodWithData, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "Title":
		return m.Title
	case "ServiceType":
		return m.ServiceType.String()
	case "StartTime":
		return m.StartTime.Format("15:04")
	case "EndTime":
		return m.EndTime.Format("15:04")
	case "Duration":
		return strconv.Itoa(m.Duration)
	case "Days":
		return m.DaysBitmask.String()
	default:
		return ""
	}
}

func periodTargetExtractor(m *PeriodWithData) string {
	return m.ID
}

func educatorsRenderer(item any) templ.Component {
	p := item.(PeriodWithData)
	if len(p.Educators) == 0 {
		return templ.NopComponent
	}
	people := make([]sharedmodels.Person, len(p.Educators))
	for i, educator := range p.Educators {
		people[i] = educator.Person
	}
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return coreblocks.PersonAvatars("educator", people).Render(ctx, w)
	})
}

func studentsRenderer(item any) templ.Component {
	p := item.(PeriodWithData)
	if len(p.Students) == 0 {
		return templ.NopComponent
	}
	people := make([]sharedmodels.Person, len(p.Students))
	for i, student := range p.Students {
		people[i] = student.Person
	}
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return coreblocks.PersonAvatars("student", people).Render(ctx, w)
	})
}
