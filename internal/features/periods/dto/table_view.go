package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/periods/models"
	"strconv"
)

var PeriodColumns = []shareddto.ColumnView{
	{Field: "Title", Display: "title"},
	{Field: "ServiceType", Display: "type", Renderer: "badge", Alignment: "center"},
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
	case "ServiceType":
		return p.ServiceType.String()
	case "StartTime":
		return p.StartTime.Format("15:04")
	case "EndTime":
		return p.EndTime.Format("15:04")
	case "Duration":
		return strconv.Itoa(p.Duration)
	case "Days":
		return p.DaysBitmask.String()
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
