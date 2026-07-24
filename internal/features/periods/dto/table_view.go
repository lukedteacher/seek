package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/periods/models"
	"strconv"
)

var PeriodColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "StartTime", Display: "start", Group: "time"},
	{Field: "EndTime", Display: "end", Group: "time"},
	{Field: "Duration", Display: "duration (min)"},
	{Field: "Days", Display: "days"},
	// id omitted from display (auto‑stored in RowView.ID)
}

func extractPeriod(s *models.Period, field string) string {
	if s == nil {
		return ""
	}
	switch field {
	case "ID":
		return s.ID
	case "Title":
		return s.Title
	case "StartTime":
		return s.StartTime
	case "Duration":
		return strconv.FormatInt(s.Duration, 10)
	case "Days":
		return strconv.FormatInt(s.Days, 10)
	default:
		return ""
	}
}

var PeriodTableConfig = shareddto.TableConfig[models.Period]{
	Name:    "periods",
	Columns: PeriodColumns,
	Extract: extractPeriod,
}

func NewPeriodTableView(periods []models.Period) shareddto.TableView {
	return shareddto.NewTableView(periods, PeriodTableConfig)
}
