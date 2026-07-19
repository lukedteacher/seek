package dto

import (
	"fmt"
	"reflect"
)

type TableView struct {
	Columns []ColumnMeta
	Rows    []RowView
}

type ColumnMeta struct {
	Display  string
	JSON     string
	Renderer string // "text", "badge" – default "text"
}

type RowView struct {
	ID    string
	Cells []CellView
}

type CellView string

// BuildTableView converts a slice of structs to a TableView.
// parameters:
//   - items: slice of structs (or pointers to structs)
//   - hideFields: field names to exclude from the visible columns (e.g., ["CaseManager"])
//   - includeFields: if non-nil, only these fields (in this order) are shown (ignored if nil)
//
// The "ID" field is always hidden and stored in RowView.ID.
func BuildTableView[T any](items []T, hideFields []string, includeFields []string) TableView {
	if len(items) == 0 {
		return TableView{Columns: []ColumnMeta{}, Rows: []RowView{}}
	}

	t := reflect.TypeOf(items[0])
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// build a map of field name -> struct field, and collect all fields
	fieldMap := make(map[string]reflect.StructField)
	allFieldNames := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldMap[f.Name] = f
		allFieldNames = append(allFieldNames, f.Name)
	}

	// determine visible fields
	var visibleFields []reflect.StructField
	if includeFields != nil {
		// only include fields from the provided list
		for _, name := range includeFields {
			if f, ok := fieldMap[name]; ok && name != "ID" && !contains(hideFields, name) {
				visibleFields = append(visibleFields, f)
			}
		}
	} else {
		// use all fields except ID and those in hideFields, preserving struct order
		for _, name := range allFieldNames {
			if name == "ID" || contains(hideFields, name) {
				continue
			}
			visibleFields = append(visibleFields, fieldMap[name])
		}
	}

	// build column metadata
	columns := make([]ColumnMeta, len(visibleFields))
	for i, f := range visibleFields {
		display := f.Tag.Get("display")
		json := f.Tag.Get("json")
		renderer := f.Tag.Get("renderer")
		if renderer == "" {
			renderer = "text"
		}
		columns[i] = ColumnMeta{
			Display:  display,
			JSON:     json,
			Renderer: renderer,
		}
	}

	// build rows
	rows := make([]RowView, len(items))
	for i, item := range items {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// Extract ID (if field exists)
		idVal := v.FieldByName("ID")
		idStr := formatValue(idVal)

		cells := make([]CellView, len(visibleFields))
		for j, f := range visibleFields {
			fieldVal := v.FieldByName(f.Name)
			// check for format tag: call method on struct
			if formatMethod := f.Tag.Get("format"); formatMethod != "" {
				method := v.MethodByName(formatMethod)
				if method.IsValid() && method.Kind() == reflect.Func {
					results := method.Call(nil)
					if len(results) == 1 && results[0].Kind() == reflect.String {
						cells[j] = CellView(results[0].String())
						continue
					}
				}
			}
			// default formatting
			cells[j] = CellView(formatValue(fieldVal))
		}

		rows[i] = RowView{ID: idStr, Cells: cells}
	}

	return TableView{Columns: columns, Rows: rows}
}

// formatValue handles pointers and basic types
func formatValue(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return ""
		}
		return fmt.Sprint(val.Elem().Interface())
	default:
		return fmt.Sprint(val.Interface())
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
