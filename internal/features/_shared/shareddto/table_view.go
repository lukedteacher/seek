package shareddto

type TableConfig[T any] struct {
	Name            string
	Columns         []ColumnView
	ValueExtractor  func(item *T, field string) string // for cell values
	TargetExtractor func(item *T) string               // for row link target
}

type TableView struct {
	Name    string
	Columns []ColumnView
	Rows    []RowView
	URL     string
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
	Target string
	Cells  []CellView
}

type CellView string

// NewTableView converts a slice of structs to a TableView.
// the "ID" field is always hidden and stored in RowView.ID.
func NewTableView[T any](items []T, cfg TableConfig[T]) TableView {
	if len(items) == 0 || len(cfg.Columns) == 0 {
		return TableView{Name: cfg.Name, Columns: cfg.Columns, Rows: []RowView{}}
	}

	rows := make([]RowView, len(items))
	for i := range items {
		item := &items[i]
		target := cfg.TargetExtractor(item)

		cells := make([]CellView, len(cfg.Columns))
		for j, col := range cfg.Columns {
			cells[j] = CellView(cfg.ValueExtractor(item, col.Field))
		}
		rows[i] = RowView{Target: target, Cells: cells}
	}

	return TableView{
		Name:    cfg.Name,
		Columns: cfg.Columns,
		Rows:    rows,
	}
}
