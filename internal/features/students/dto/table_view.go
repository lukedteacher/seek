package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/students/models"
)

var StudentColumns = []shareddto.ColumnView{
	{Field: "FirstName", Display: "given", Group: "name"},
	{Field: "ChosenName", Display: "chosen", Group: "name"},
	{Field: "LastName", Display: "family", Group: "name"},
	{Field: "Grade", Display: "grade", Renderer: "badge"},
	{Field: "Homeroom", Display: "homeroom"},
	{Field: "CaseManager", Display: "case manager"},
	// id omitted from display (auto‑stored in RowView.ID)
}

func extractStudent(s *models.Student, field string) string {
	if s == nil {
		return ""
	}
	switch field {
	case "ID":
		return s.ID
	case "FirstName":
		return s.GivenName
	case "ChosenName":
		return s.ChosenName
	case "LastName":
		return s.FamilyName
	case "Grade":
		return s.GradeOrdinal() // method on Student
	case "Homeroom":
		return s.Homeroom
	case "CaseManager":
		return s.CaseManager
	default:
		return ""
	}
}

var StudentTableConfig = shareddto.TableConfig[models.Student]{
	Name:    "students",
	Columns: StudentColumns,
	Extract: extractStudent,
}

func NewStudentTableView(students []models.Student) shareddto.TableView {
	return shareddto.NewTableView(students, StudentTableConfig)
}
