package dto

import (
	"strconv"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/iep_services/models"
)

// columns for the iep service table
var IEPServiceColumns = []shareddto.ColumnView{
	{Field: "ServiceName", Display: "name"},
	{Field: "ServiceType", Display: "type", Renderer: "badge", Alignment: "center"},
	{Field: "IndirectMinutes", Display: "indirect", Group: "minutes"},
	{Field: "DirectMinutes", Display: "direct", Group: "minutes"},
	{Field: "FrequencyCount", Display: "count", Group: "frequency"},
	{Field: "FrequencyType", Display: "type", Group: "frequency", Renderer: "badge", Alignment: "center"},
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
	case "ServiceName":
		return svc.ServiceName
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
		return svc.StartDate.Format("02 Jan, 2006")
	case "EndDate":
		return svc.EndDate.Format("02 Jan, 2006")
	case "CreatedAt":
		return svc.CreatedAt.Format("02 Jan, 2006")
	case "UpdatedAt":
		return svc.UpdatedAt.Format("02 Jan, 2006")
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
