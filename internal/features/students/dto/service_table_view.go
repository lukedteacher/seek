package dto

import (
	"strconv"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/services/models"
)

// columns for the iep service table
var ServiceColumns = []shareddto.ColumnView{
	{Field: "ServiceName", Display: "name"},
	{Field: "ServiceType", Display: "type", Renderer: "badge", Alignment: "center"},
	{Field: "IndirectMinutes", Display: "indirect", Group: "minutes", Alignment: "center"},
	{Field: "DirectMinutes", Display: "direct", Group: "minutes", Alignment: "center"},
	{Field: "FrequencyCount", Display: "count", Group: "frequency", Alignment: "center"},
	{Field: "FrequencyType", Display: "type", Group: "frequency", Renderer: "badge", Alignment: "center"},
	{Field: "Location", Display: "location", Renderer: "badge", Alignment: "center"},
	{Field: "Provider", Display: "provider"},
	{Field: "StartDate", Display: "start", Group: "date", Alignment: "center"},
	{Field: "EndDate", Display: "end", Group: "date", Alignment: "center"},
	// id, created_at, updated_at, archived_at omitted from display
}

// extract values from an iep service by field name
func serviceValueExtractor(m *models.Service, field string) string {
	if m == nil {
		return ""
	}
	switch field {
	case "ID":
		return m.ID
	case "StudentID":
		return m.StudentID
	case "ServiceName":
		return m.ServiceName
	case "ServiceType":
		return m.ServiceType.ShortString()
	case "IndirectMinutes":
		return strconv.Itoa(m.IndirectMinutes)
	case "DirectMinutes":
		return strconv.Itoa(m.DirectMinutes)
	case "FrequencyCount":
		return strconv.Itoa(m.FrequencyCount)
	case "FrequencyType":
		return m.FrequencyType
	case "LocationID":
		return m.LocationID
	case "Provider":
		return m.Provider
	case "StartDate":
		return m.StartDate.String()
	case "EndDate":
		return m.EndDate.String()
	case "CreatedAt":
		return m.CreatedAt.Format("02 Jan, 2006")
	case "UpdatedAt":
		return m.UpdatedAt.Format("02 Jan, 2006")
	default:
		return ""
	}
}

func serviceTargetExtractor(m *models.Service) string {
	return m.ID
}

// table config for IEP service (used by both regular and diff tables)
var ServiceTableConfig = shareddto.TableConfig[models.Service]{
	Name:            "services",
	Columns:         ServiceColumns,
	ValueExtractor:  serviceValueExtractor,
	TargetExtractor: serviceTargetExtractor,
}

func NewServiceTableView(services []models.Service) shareddto.TableView {
	return shareddto.NewTableView(services, ServiceTableConfig)
}
