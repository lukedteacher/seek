package dto

import (
	"strings"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/educators/models"
)

// columns for the educator table
var EducatorColumns = []shareddto.ColumnView{
	{Field: "GivenName", Display: "given", Group: "name", Signal: "given_name"},
	{Field: "ChosenName", Display: "chosen", Group: "name", Signal: "chosen_name"},
	{Field: "FamilyName", Display: "family", Group: "name", Signal: "family_name"},
	{Field: "Roles", Display: "role(s)", Renderer: "badge", Alignment: "center", Signal: "roles"},
}

// extract values from an educator by field name
func valueExtractor(m *models.Educator, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "GivenName":
		return m.GivenName
	case "ChosenName":
		return m.ChosenName
	case "FamilyName":
		return m.FamilyName
	case "Email":
		return m.Email
	case "Roles":
		var parts []string
		for _, r := range m.Roles {
			parts = append(parts, r.String())
		}
		return strings.Join(parts, ",")
	default:
		return ""
	}
}

// uses username as the target for the resource URL
func targetExtractor(m *models.Educator) string {
	return m.Username
}

// table config for educator (used by both regular and diff tables)
var EducatorTableConfig = shareddto.TableConfig[models.Educator]{
	Name:    "educators",
	Columns: EducatorColumns,
	Sort: shareddto.TableSort{
		Column:    "family_name",
		Direction: "ASC",
	},
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

// convenience constructor for a regular table view of educators
func NewEducatorTableView(educators []models.Educator) shareddto.TableView {
	return shareddto.NewTableView(educators, EducatorTableConfig)
}

func NewEducatorTableViewWithSort(educators []models.Educator, sort shareddto.TableSort) shareddto.TableView {
	if sort.Column != "" && sort.Direction != "" {
		EducatorTableConfig.Sort = sort
	}
	return shareddto.NewTableView(educators, EducatorTableConfig)
}
