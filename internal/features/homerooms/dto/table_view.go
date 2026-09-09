package dto

import (
	"context"
	"io"
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/homerooms/models"
	"seek/internal/ui/core/coreblocks"
	"strconv"

	"github.com/a-h/templ"
)

var HomeroomColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "Grades", Display: "grades", RenderFunc: gradesRenderer},
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

func gradesRenderer(item any) templ.Component {
	s := item.(models.Homeroom)
	if s.GradesBitmask == -1 {
		return templ.NopComponent
	}
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		component := coreblocks.GradeBadges(s.GradesBitmask)
		return component.Render(ctx, w)
	})
}
