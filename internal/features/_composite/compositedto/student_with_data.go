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

type Student struct {
	studentModels.Student
	CaseManager educatorModels.Educator
	Services    []serviceModels.IEPService
}

var StudentColumns = []shareddto.ColumnView{
	{Field: "MARSSID", Display: "MARSS ID"},
	{Field: "GivenName", Display: "given", Group: "name"},
	{Field: "ChosenName", Display: "chosen", Group: "name"},
	{Field: "FamilyName", Display: "family", Group: "name"},
	{Field: "Email", Display: "email"},
	{Field: "Grade", Display: "grade", Renderer: "badge", Alignment: "center"},
	{Field: "Homeroom", Display: "homeroom"},
	{Field: "CaseManager", Display: "case manager", RenderFunc: caseManagerRenderer},
}

var StudentTableConfig = shareddto.TableConfig[Student]{
	Name:            "students",
	Columns:         StudentColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
	SubTableBuilder: func(student Student) shareddto.TableView {
		return serviceDTO.NewIEPServiceTableView(student.Services)
	},
}

func NewStudentTableView(students []Student) shareddto.TableView {
	return shareddto.NewTableView(students, StudentTableConfig)
}

func valueExtractor(m *Student, field string) string {
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

func targetExtractor(m *Student) string {
	return m.Username
}

func caseManagerRenderer(item any) templ.Component {
	s := item.(Student)
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
