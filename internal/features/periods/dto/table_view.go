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
	{Field: "Duration", Display: "duration", Alignment: "center"},
	{Field: "Days", Display: "days", Alignment: "center"},
	{Field: "EducatorIDs", Display: "educators", Alignment: "center"},
	{Field: "StudentIDs", Display: "students", Alignment: "center"},
}

func valueExtractor(m *models.Period, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "Title":
		return m.Title
	case "ServiceType":
		return m.ServiceType.String()
	case "StartTime":
		return m.StartTime.Format("15:04")
	case "EndTime":
		return m.EndTime.Format("15:04")
	case "Duration":
		return strconv.Itoa(m.Duration)
	case "Days":
		return m.DaysBitmask.String()
	case "EducatorIDs":
		return strconv.Itoa(len(m.EducatorIDs))
	case "StudentIDs":
		return strconv.Itoa(len(m.StudentIDs))
	default:
		return ""
	}
}

func targetExtractor(m *models.Period) string {
	return m.ID
}

var PeriodTableConfig = shareddto.TableConfig[models.Period]{
	Name:            "periods",
	Columns:         PeriodColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

func NewPeriodTableView(periods []models.Period) shareddto.TableView {
	return shareddto.NewTableView(periods, PeriodTableConfig)
}
