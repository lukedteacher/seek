package dto

import (
	"fmt"
	"strings"
	"time"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/educators/models"
)

// columns for the educator table
var EducatorTableColumns = []shareddto.ColumnView{
	{Field: "FirstL", Display: "first l", Group: "name", Signal: "first_l", Sortable: true, Groupable: false, Visible: true},
	{Field: "GivenName", Display: "given", Group: "name", Signal: "given_name", Sortable: true, Groupable: true, Visible: true},
	{Field: "ChosenName", Display: "chosen", Group: "name", Signal: "chosen_name", Sortable: true, Groupable: false, Visible: true},
	{Field: "FamilyName", Display: "family", Group: "name", Signal: "family_name", Sortable: true, Groupable: true, Visible: true},
	{Field: "Email", Display: "email", Signal: "email", Sortable: true, Groupable: false, Visible: true},
	{Field: "Roles", Display: "role(s)", Renderer: "badge", Alignment: "center", Signal: "roles", Sortable: false, Groupable: true, Visible: true},
	{
		Field:    "UpdatedAt",
		Signal:   "updated_at",
		Display:  "last updated",
		Sortable: true,
		Visible:  true,
	},
}

// extract values from an educator by field name
func valueExtractor(m *models.Educator, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "FirstL":
		return m.NameInitial()
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
	case "UpdatedAt":
		return relativeTime(m.UpdatedAt)
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
	Columns: EducatorTableColumns,
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

func NewEducatorTableViewWithSort(
	educators []models.Educator,
	columns map[string]bool,
	sort shareddto.TableSort,
	groupedBy string,
) shareddto.TableView {
	if sort.Column != "" && sort.Direction != "" {
		EducatorTableConfig.Sort = sort
	}

	visible := make([]shareddto.ColumnView, 0, len(EducatorTableColumns))
	for _, c := range EducatorTableColumns {
		if columns[c.Signal] {
			visible = append(visible, c)
		}
	}
	EducatorTableConfig.Columns = visible

	EducatorTableConfig.GroupExtractor = educatorGroupExtractor(groupedBy)

	return shareddto.NewTableView(educators, EducatorTableConfig)
}

func educatorGroupExtractor(groupedBy string) func(*models.Educator) string {
	switch groupedBy {
	case "given_name":
		return func(e *models.Educator) string {
			if e.GivenName == "" {
				return ""
			}
			return strings.ToUpper(e.GivenName[:1])
		}
	case "family_name":
		return func(e *models.Educator) string {
			if e.FamilyName == "" {
				return ""
			}
			return strings.ToUpper(e.FamilyName[:1])
		}
	case "role":
		return func(e *models.Educator) string {
			if len(e.Roles) == 0 {
				return ""
			}
			return string(e.Roles[0])
		}
	default:
		return nil // "none" → no grouping
	}
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
