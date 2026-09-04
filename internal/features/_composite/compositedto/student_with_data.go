package compositedto

import (
	"context"
	"io"
	"seek/internal/features/_shared/shareddto"
	educatorBlocks "seek/internal/features/educators/blocks"
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	iepModels "seek/internal/features/ieps/models"
	serviceDTO "seek/internal/features/services/dto"
	serviceModels "seek/internal/features/services/models"
	studentModels "seek/internal/features/students/models"

	"github.com/a-h/templ"
)

type StudentWithData struct {
	studentModels.Student
	IEP         iepModels.IEP
	CaseManager educatorModels.Educator
	Services    []serviceModels.Service
}

var StudentWithDataTableConfig = shareddto.TableConfig[StudentWithData]{
	Name: "students",
	Sort: shareddto.TableSort{
		Column:    "family_name",
		Direction: "ASC",
	},
	Columns:         StudentWithDataColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
	SubTableBuilder: func(student StudentWithData) shareddto.TableView {
		return serviceDTO.NewServiceTableView(student.Services)
	},
}

var StudentWithDataColumns = []shareddto.ColumnView{
	{Field: "GivenName", Display: "given", Group: "name", Signal: "given_name"},
	{Field: "ChosenName", Display: "chosen", Group: "name", Signal: "chosen_name"},
	{Field: "FamilyName", Display: "family", Group: "name", Signal: "family_name"},
	{Field: "Email", Display: "email", Signal: "email"},
	{Field: "Grade", Display: "grade", Renderer: "badge", Alignment: "center", Signal: "grade"},
	{Field: "Homeroom", Display: "homeroom", Signal: "homeroom"},
	{Field: "PlanType", Display: "plan", Renderer: "badge", Alignment: "center", Signal: "plan_type"},
	{Field: "CaseManager", Display: "case manager", RenderFunc: caseManagerRenderer, Signal: "case_manager"},
}

func NewStudentWithDataTableView(students []StudentWithData, sort shareddto.TableSort) shareddto.TableView {
	if sort.Column != "" && sort.Direction != "" {
		StudentWithDataTableConfig.Sort = sort
	}
	return shareddto.NewTableView(students, StudentWithDataTableConfig)
}

func valueExtractor(m *StudentWithData, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "MARSSID":
		return m.MARSSID
	case "GivenName":
		return m.GivenName
	case "ChosenName":
		return m.ChosenName
	case "FamilyName":
		return m.FamilyName
	case "Email":
		return m.Email
	case "Grade":
		return m.Grade.Ordinal()
	case "Homeroom":
		if m.HomeroomID != "" {
			return m.HomeroomID
		}
		return ""
	case "PlanType":
		return m.PlanType.Description()
	default:
		return ""
	}
}

func targetExtractor(m *StudentWithData) string {
	return m.Username
}

func caseManagerRenderer(item any) templ.Component {
	s := item.(StudentWithData)
	if s.CaseManager.ID == "" {
		return templ.NopComponent
	}
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		educatorView := educatorDTO.EducatorView{
			Person: s.CaseManager.Person,
			ID:     s.CaseManagerID,
		}
		return educatorBlocks.EducatorAvatar(educatorView).Render(ctx, w)
	})
}
