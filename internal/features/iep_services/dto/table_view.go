package dto

import (
	"strconv"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/iep_services/models"
)

// columns for the iep service table
var IEPServiceColumns = []shareddto.ColumnView{
	{Field: "ServiceType", Display: "type", Renderer: "badge", Alignment: "center"},
	{Field: "IndirectMinutes", Display: "indirect", Group: "minutes"},
	{Field: "DirectMinutes", Display: "direct", Group: "minutes"},
	{Field: "FrequencyCount", Display: "count", Group: "frequency"},
	{Field: "FrequencyType", Display: "type", Group: "frequency"},
	{Field: "Location", Display: "location", Renderer: "badge", Alignment: "center"},
	{Field: "Provider", Display: "provider"},
	{Field: "StartDate", Display: "start", Group: "date"},
	{Field: "EndDate", Display: "end", Group: "date"},
	// id, created_at, updated_at, archived_at omitted from display
}

// extract values from an iep service by field name
func extractIEPService(svc *models.IEPService, field string) string {
	if svc == nil {
		return ""
	}
	switch field {
	case "ID":
		return svc.ID
	case "StudentID":
		return svc.StudentID
	case "ServiceType":
		return svc.ServiceType.ShortString()
	case "IndirectMinutes":
		return strconv.Itoa(svc.IndirectMinutes)
	case "DirectMinutes":
		return strconv.Itoa(svc.DirectMinutes)
	case "FrequencyCount":
		return strconv.Itoa(svc.FrequencyCount)
	case "FrequencyType":
		return svc.FrequencyType
	case "Location":
		return svc.Location
	case "Provider":
		return svc.Provider
	case "StartDate":
		return svc.StartDate
	case "EndDate":
		return svc.EndDate
	case "CreatedAt":
		return svc.CreatedAt
	case "UpdatedAt":
		return svc.UpdatedAt
	case "ArchivedAt":
		return svc.ArchivedAt
	default:
		return ""
	}
}

// table config for IEP service (used by both regular and diff tables)
var IEPServiceTableConfig = shareddto.TableConfig[models.IEPService]{
	Columns: IEPServiceColumns,
	Extract: extractIEPService,
}

func NewIEPServiceTableView(services []models.IEPService) shareddto.TableView {
	return shareddto.NewTableView(services, IEPServiceTableConfig)
}
