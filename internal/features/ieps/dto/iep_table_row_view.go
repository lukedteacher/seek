package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/ieps/models"
)

// columns for the iep service table
var IEPColumns = []shareddto.ColumnView{
	{Field: "StudentID", Display: "student ID"},
	{Field: "StartDate", Display: "start", Group: "date", Alignment: "center"},
	{Field: "EndDate", Display: "end", Group: "date", Alignment: "center"},
	{Field: "AmendedDate", Display: "amended", Group: "date", Alignment: "center"},
}

// extract values from an iep service by field name
func valueExtractor(m *models.IEP, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "StudentID":
		return m.StudentID
	case "StartDate":
		return m.StartDate.String()
	case "EndDate":
		return m.EndDate.String()
	case "AmendedDate":
		return m.AmendedDate.String()
	default:
		return ""
	}
}

func targetExtractor(m *models.IEP) string {
	return m.ID
}

// table config for IEP service (used by both regular and diff tables)
var IEPTableConfig = shareddto.TableConfig[models.IEP]{
	Name:            "ieps",
	Columns:         IEPColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

func NewIEPTableView(services []models.IEP) shareddto.TableView {
	return shareddto.NewTableView(services, IEPTableConfig)
}
