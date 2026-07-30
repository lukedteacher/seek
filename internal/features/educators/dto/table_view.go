package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/educators/models"
)

// columns for the educator table
var EducatorColumns = []shareddto.ColumnView{
	{Field: "GivenName", Display: "given", Group: "name"},
	{Field: "ChosenName", Display: "chosen", Group: "name"},
	{Field: "FamilyName", Display: "family", Group: "name"},
	{Field: "Email", Display: "email"},
	{Field: "Role", Display: "role", Renderer: "badge", Alignment: "center"},
	// id omitted from display (auto‑stored in RowView.ID)
}

// extract values from an educator by field name
func valueExtractor(svc *models.Educator, field string) string {
	if svc == nil {
		return ""
	}
	switch field {
	case "ID":
		return svc.ID
	case "GivenName":
		return svc.GivenName
	case "ChosenName":
		return svc.ChosenName
	case "FamilyName":
		return svc.FamilyName
	case "Email":
		return svc.Email
	case "Role":
		return svc.Role
	default:
		return ""
	}
}

func targetExtractor(m *models.Educator) string {
	return m.ID
}

// table config for educator (used by both regular and diff tables)
var EducatorTableConfig = shareddto.TableConfig[models.Educator]{
	Name:            "educators",
	Columns:         EducatorColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

// convenience constructor for a regular table view of educators
func NewEducatorTableView(educators []models.Educator) shareddto.TableView {
	return shareddto.NewTableView(educators, EducatorTableConfig)
}
