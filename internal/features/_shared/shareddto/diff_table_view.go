package shareddto

import (
	"seek/internal/features/_shared/sharedmodels"
)

type DiffTableView struct {
	Name    string
	Columns []ColumnView
	Rows    []DiffRowView
}

type DiffRowView struct {
	ID       string
	Status   sharedmodels.DiffStatus // "same, "updated", "added", "deleted"
	OldCells []CellView              // values from DB (nil for added)
	NewCells []CellView              // values from CSV (nil for deleted)
}

func NewDiffTableView[T any](diffs []sharedmodels.Diff[T], cfg TableConfig[T]) DiffTableView {
	rows := make([]DiffRowView, len(diffs))
	for i, d := range diffs {
		rows[i] = DiffRowView{
			ID:       d.Key,
			Status:   d.Status,
			OldCells: extractCells(d.Old, cfg.Columns, cfg.ValueExtractor),
			NewCells: extractCells(d.New, cfg.Columns, cfg.ValueExtractor),
		}
	}
	return DiffTableView{
		Name:    cfg.Name,
		Columns: cfg.Columns,
		Rows:    rows,
	}
}

func extractCells[T any](
	ptr *T,
	cols []ColumnView,
	extract func(*T, string) string,
) []CellView {
	if ptr == nil {
		return make([]CellView, len(cols))
	}
	vals := make([]CellView, len(cols))
	for i, col := range cols {
		vals[i] = CellView(extract(ptr, col.Field))
	}
	return vals
}

func uniqueGroups(cols []ColumnView) []string {
	seen := map[string]bool{}
	var groups []string
	for _, c := range cols {
		if c.Group == "" {
			continue
		}
		if !seen[c.Group] {
			seen[c.Group] = true
			groups = append(groups, c.Group)
		}
	}
	return groups
}

func countColumnsInGroup(cols []ColumnView, group string) int {
	count := 0
	for _, c := range cols {
		if c.Group == group {
			count++
		}
	}
	return count
}
