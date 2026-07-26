package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/periods/models"
	"strconv"
)

var PeriodColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "StartTime", Display: "start", Group: "time", Alignment: "center"},
	{Field: "EndTime", Display: "end", Group: "time", Alignment: "center"},
	{Field: "Duration", Display: "duration (min)", Alignment: "center"},
	{Field: "Days", Display: "days", Alignment: "center"},
	// id omitted from display (auto‑stored in RowView.ID)
}

func extractPeriod(p *models.Period, field string) string {
	if p == nil {
		return ""
	}
	switch field {
	case "ID":
		return p.ID
	case "Title":
		return p.Title
	case "StartTime":
		return p.StartTime.Format("15:04")
	case "EndTime":
		return p.EndTime.Format("15:04")
	case "Duration":
		return strconv.FormatInt(p.Duration, 10)
	case "Days":
		return strconv.FormatInt(p.Days, 10)
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
