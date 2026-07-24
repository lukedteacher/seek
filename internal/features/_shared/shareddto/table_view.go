package shareddto

import (
	"reflect"
)

type TableConfig[T any] struct {
	Name    string
	Columns []ColumnView
	Extract func(*T, string) string
}

type TableView struct {
	Name    string
	Columns []ColumnView
	Rows    []RowView
}

type ColumnView struct {
	Field        string // e.g. "DirectMinutes", "StudentID"
	JSON         string // e.g. "direct_minutes", "student_id"
	Display      string // e.g. "direct", "student ID"
	Group        string // e.g. "minutes", ""
	Renderer     string // e.g. "text", "badge"
	Alignment    string // e.g. "left", "center", "right"
	FormatMethod string // e.g. "GradeOrdinal", "FullName"
}

type RowView struct {
	ID    string
	Cells []CellView
}

type CellView string

// NewTableView converts a slice of structs to a TableView.
// the "ID" field is always hidden and stored in RowView.ID.
func NewTableView[T any](items []T, cfg TableConfig[T]) TableView {
	if len(items) == 0 || len(cfg.Columns) == 0 {
		return TableView{Name: cfg.Name, Columns: cfg.Columns, Rows: []RowView{}}
	}

	rows := make([]RowView, len(items))
	for i, item := range items {
		// id extraction using reflection
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		idVal := v.FieldByName("ID")

		cells := make([]CellView, len(cfg.Columns))
		for j, col := range cfg.Columns {
			cells[j] = CellView(cfg.Extract(&item, col.Field))
		}
		rows[i] = RowView{ID: idVal.String(), Cells: cells}
	}

	return TableView{
		Name:    cfg.Name,
		Columns: cfg.Columns,
		Rows:    rows,
	}
}
