package shareddto

import "github.com/a-h/templ"

type TableConfig[T any] struct {
	Name            string
	Sort            TableSort
	Columns         []ColumnView
	ValueExtractor  func(item *T, field string) string // for cell values
	TargetExtractor func(item *T) string               // for row link target
	SubTableBuilder func(item T) TableView
}

type TableSort struct {
	Column    string
	Direction string
}

type TableView struct {
	Name    string
	Sort    TableSort
	Columns []ColumnView
	Rows    []RowView
}

type ColumnView struct {
	Field        string                         // e.g. "DirectMinutes", "StudentID"
	Signal       string                         // e.g. "direct_minutes", "student_id"
	Display      string                         // e.g. "direct", "student ID"
	Group        string                         // e.g. "minutes", ""
	Renderer     string                         // e.g. "text", "badge"
	Alignment    string                         // e.g. "left", "center", "right"
	FormatMethod string                         // e.g. "GradeOrdinal", "FullName"
	RenderFunc   func(item any) templ.Component // optional custom cell renderer
}

type RowView struct {
	Target   string
	Item     any
	Cells    []CellView
	SubTable *TableView
}

type CellView string

func (cv CellView) String() string {
	return string(cv)
}

// NewTableView converts a slice of structs to a TableView.
// the ID field is always hidden
func NewTableView[T any](items []T, cfg TableConfig[T]) TableView {
	if len(items) == 0 || len(cfg.Columns) == 0 {
		return TableView{
			Name:    cfg.Name,
			Sort:    cfg.Sort,
			Columns: cfg.Columns,
			Rows:    []RowView{},
		}
	}

	rows := make([]RowView, len(items))
	for i := range items {
		item := &items[i]
		target := cfg.TargetExtractor(item)

		cells := make([]CellView, len(cfg.Columns))
		for j, col := range cfg.Columns {
			cells[j] = CellView(cfg.ValueExtractor(item, col.Field))
		}

		var subTable *TableView
		if cfg.SubTableBuilder != nil {
			st := cfg.SubTableBuilder(*item)
			subTable = &st
		}

		rows[i] = RowView{
			Target:   target,
			Item:     *item,
			Cells:    cells,
			SubTable: subTable,
		}
	}

	return TableView{
		Name:    cfg.Name,
		Sort:    cfg.Sort,
		Columns: cfg.Columns,
		Rows:    rows,
	}
}
