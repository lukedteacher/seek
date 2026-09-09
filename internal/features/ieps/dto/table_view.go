package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/ieps/models"
	"strconv"
)

// table config (used by both regular and diff tables)
var IEPTableConfig = shareddto.TableConfig[models.IEP]{
	Name:            "ieps",
	Columns:         IEPColumns,
	ValueExtractor:  valueExtractor,
	TargetExtractor: targetExtractor,
}

// columns for the iep service table
var IEPColumns = []shareddto.ColumnView{
	{Field: "StudentMARSSID", Display: "student MARSS ID"},
	{Field: "PlanManagerSPEDFormsID", Display: "pm sf id"},
	{Field: "Disability1", Display: "disability 1", Renderer: "badge", Alignment: "center"},
	{Field: "Disability2", Display: "disability 2", Renderer: "badge", Alignment: "center"},
	{Field: "MeetingDate", Display: "meeting", Group: "date", Alignment: "center"},
	{Field: "IEPDueDate", Display: "IEP due", Group: "date", Alignment: "center"},
	{Field: "LastEvalDate", Display: "last eval", Group: "date", Alignment: "center"},
	{Field: "EvalDueDate", Display: "eval due", Group: "date", Alignment: "center"},
	{Field: "AmendedDate", Display: "amended", Group: "date", Alignment: "center"},
	{Field: "IEPType", Display: "IEP type", Renderer: "badge", Alignment: "center"},
}

func NewIEPTableView(services []models.IEP) shareddto.TableView {
	return shareddto.NewTableView(services, IEPTableConfig)
}

// extract values from an iep service by field name
func valueExtractor(m *models.IEP, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "StudentMARSSID":
		return m.StudentMARSSID
	case "PlanManagerSPEDFormsID":
		return m.PlanManagerSPEDFormsID
	case "Disability1":
		if s := m.Disability1.Short(); s != "ND" {
			return s
		}
		return ""
	case "Disability2":
		if s := m.Disability2.Short(); s != "ND" {
			return s
		}
		return ""
	case "FederalSetting":
		return strconv.Itoa(m.FederalSetting)
	case "MeetingDate":
		return m.MeetingDate.String()
	case "IEPDueDate":
		return m.IEPDueDate.String()
	case "LastEvalDate":
		return m.LastEvalDate.String()
	case "EvalDueDate":
		return m.EvalDueDate.String()
	case "AmendedDate":
		if s := m.AmendedDate.String(); s != "" {
			return s
		}
		return "-"
	case "IEPType":
		return m.IEPType.Word()
	default:
		return ""
	}
}

func targetExtractor(m *models.IEP) string {
	return m.ID
}
