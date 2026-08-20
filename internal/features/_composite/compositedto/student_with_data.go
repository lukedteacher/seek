package compositedto

import (
	"context"
	"io"
	"seek/internal/features/_shared/shareddto"
	educatorBlocks "seek/internal/features/educators/blocks"
	educatorDTO "seek/internal/features/educators/dto"
	educatorModels "seek/internal/features/educators/models"
	serviceDTO "seek/internal/features/iepservices/dto"
	serviceModels "seek/internal/features/iepservices/models"
	studentModels "seek/internal/features/students/models"

	"github.com/a-h/templ"
)

type StudentWithData struct {
	studentModels.Student
	CaseManager educatorModels.Educator
	Services    []serviceModels.IEPService
}

var StudentWithDataColumns = []shareddto.ColumnView{
	{Field: "GivenName", Display: "given", Group: "name", Signal: "given_name"},
	{Field: "ChosenName", Display: "chosen", Group: "name", Signal: "chosen_name"},
	{Field: "FamilyName", Display: "family", Group: "name", Signal: "family_name"},
	{Field: "Email", Display: "email", Signal: "email"},
	{Field: "Grade", Display: "grade", Renderer: "badge", Alignment: "center", Signal: "grade"},
	{Field: "Homeroom", Display: "homeroom", Signal: "homeroom"},
	{Field: "CaseManager", Display: "case manager", RenderFunc: caseManagerRenderer, Signal: "case_manager"},
}

var StudentWithDataTableConfig = shareddto.TableConfig[StudentWithData]{
	Name:            "students",
	Columns:         StudentWithDataColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
	SubTableBuilder: func(student StudentWithData) shareddto.TableView {
		return serviceDTO.NewIEPServiceTableView(student.Services)
	},
}

func NewStudentWithDataTableView(students []StudentWithData) shareddto.TableView {
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
		if m.Homeroom != "" {
			return m.Homeroom
		}
		return ""
	case "CaseManager":
		if m.CaseManagerID != "" {
			return m.CaseManager.NameInitial()
		}
		return ""
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
