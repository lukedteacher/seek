package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/students/models"
)

var StudentColumns = []shareddto.ColumnView{
	{Field: "MARSSID", Display: "MARSS ID"},
	{Field: "GivenName", Display: "given", Group: "name"},
	{Field: "ChosenName", Display: "chosen", Group: "name"},
	{Field: "FamilyName", Display: "family", Group: "name"},
	{Field: "Email", Display: "email"},
	{Field: "Grade", Display: "grade", Renderer: "badge", Alignment: "center"},
	{Field: "Homeroom", Display: "homeroom"},
	{Field: "CaseManager", Display: "case manager"},
	// id omitted from display (auto‑stored in RowView.ID)
}

func valueExtractor(m *models.Student, field string) string {
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
		return m.Homeroom
	case "CaseManager":
		return m.CaseManager
	default:
		return ""
	}
}

func targetExtractor(m *models.Student) string {
	return m.Username
}

var StudentTableConfig = shareddto.TableConfig[models.Student]{
	Name:            "students",
	Columns:         StudentColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

func NewStudentTableView(students []models.Student) shareddto.TableView {
	return shareddto.NewTableView(students, StudentTableConfig)
}
