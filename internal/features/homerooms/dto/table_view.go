package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/homerooms/models"
	"strconv"
)

var HomeroomColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "LocationID", Display: "location id"},
	{Field: "EducatorIDs", Display: "educators", Alignment: "center"},
	{Field: "StudentIDs", Display: "students", Alignment: "center"},
}

func valueExtractor(m *models.Homeroom, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "Title":
		return m.Title
	case "EducatorIDs":
		return strconv.Itoa(len(m.EducatorIDs))
	case "StudentIDs":
		return strconv.Itoa(len(m.StudentIDs))
	default:
		return ""
	}
}

func targetExtractor(m *models.Homeroom) string {
	return m.ID
}

var HomeroomTableConfig = shareddto.TableConfig[models.Homeroom]{
	Name:            "homerooms",
	Columns:         HomeroomColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

func NewHomeroomTableView(homerooms []models.Homeroom) shareddto.TableView {
	return shareddto.NewTableView(homerooms, HomeroomTableConfig)
}
