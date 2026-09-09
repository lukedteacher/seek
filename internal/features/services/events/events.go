package events

import (
	"time"

	"seek/internal/appdb"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/services/models"
)

type eventType = eventstore.EventType

// event types
const (
	EventServiceAddedToIEP     eventType = "service_added_to_iep_event"
	EventServiceRemovedFromIEP eventType = "service_removed_from_iep_event"
	EventServiceUpdated        eventType = "iep_service_updated_event"
	EventServiceArchived       eventType = "iep_service_archived_event"
	EventServiceDeleted        eventType = "iep_service_deleted_event"
)

// event fields for event IDs
const (
	FieldServiceAddedToIEPEventID = "iep_service_added_to_student_event_id"
	FieldServiceUpdatedEventID    = "iep_service_updated_event_id"
	FieldServiceDeletedEventID    = "iep_service_deleted_event_id"
)

// event fields
const (
	FieldServiceID              = "iep_service_id"
	FieldServiceIEPID           = "iep_id"
	FieldServiceStudentID       = "student_id"
	FieldServiceServiceName     = "service_name"
	FieldServiceServiceType     = "service_type"
	FieldServiceIndirectMinutes = "indirect_minutes"
	FieldServiceDirectMinutes   = "direct_minutes"
	FieldServiceFrequencyCount  = "frequency_count"
	FieldServiceFrequencyType   = "frequency_type"
	FieldServiceLocationID      = "location_id"
	FieldServiceStartDate       = "start_date"
	FieldServiceEndDate         = "end_date"
	FieldServiceProviderID      = "provider_id"
	FieldServiceAddedAt         = "added_at"
	FieldServiceUpdatedAt       = "updated_at"
	FieldServiceArchivedAt      = "archived_at"
	FieldServiceDeletedAt       = "deleted_at"
	FieldServiceScopeID         = "scope.iep_service_added_to_student_event_id"
)

type ServiceState struct {
	ID              string `json:"id"`
	IEPID           string `json:"iep_id"`
	ServiceName     string `json:"service_name,omitempty"`
	ServiceType     string `json:"service_type,omitempty"`
	IndirectMinutes int64  `json:"indirect_minutes,omitempty"`
	DirectMinutes   int64  `json:"direct_minutes,omitempty"`
	FrequencyCount  int64  `json:"frequency_count,omitempty"`
	FrequencyType   string `json:"frequency_type,omitempty"`
	LocationID      string `json:"location_id,omitempty"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
	ProviderID      string `json:"provider_id,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
	ArchivedAt      string `json:"archived_at,omitempty"`
	DeletedAt       string `json:"deleted_at,omitempty"`
}

func NewStateFromModel(m models.Service) ServiceState {
	return ServiceState{
		ID:              m.ID,
		IEPID:           m.IEPID,
		ServiceName:     m.ServiceName,
		ServiceType:     m.ServiceType.ShortString(),
		IndirectMinutes: int64(m.IndirectMinutes),
		DirectMinutes:   int64(m.DirectMinutes),
		FrequencyCount:  int64(m.FrequencyCount),
		FrequencyType:   m.FrequencyType,
		LocationID:      m.LocationID,
		StartDate:       m.StartDate.String(),
		EndDate:         m.EndDate.String(),
		ProviderID:      m.ProviderID,
		CreatedAt:       appdb.SQLTime(m.CreatedAt),
		UpdatedAt:       appdb.SQLTime(m.UpdatedAt),
	}
}

type ServiceFlat struct {
	ID              string `json:"service.id"`
	IEPID           string `json:"service.iep_id"`
	ServiceName     string `json:"service.service_name,omitempty"`
	ServiceType     string `json:"service.service_type,omitempty"`
	IndirectMinutes int64  `json:"service.indirect_minutes,omitempty"`
	DirectMinutes   int64  `json:"service.direct_minutes,omitempty"`
	FrequencyCount  int64  `json:"service.frequency_count,omitempty"`
	FrequencyType   string `json:"service.frequency_type,omitempty"`
	LocationID      string `json:"service.location_id,omitempty"`
	StartDate       string `json:"service.start_date,omitempty"`
	EndDate         string `json:"service.end_date,omitempty"`
	ProviderID      string `json:"service.provider_id,omitempty"`
	CreatedAt       string `json:"service.created_at,omitempty"`
	UpdatedAt       string `json:"service.updated_at,omitempty"`
	ArchivedAt      string `json:"service.archived_at,omitempty"`
	DeletedAt       string `json:"service.deleted_at,omitempty"`
}

func NewModelFromFlat(f ServiceFlat) models.Service {
	return models.Service{
		ID:              f.ID,
		IEPID:           f.IEPID,
		ServiceName:     f.ServiceName,
		ServiceType:     sharedmodels.ServiceType(f.ServiceType),
		IndirectMinutes: int(f.IndirectMinutes),
		DirectMinutes:   int(f.DirectMinutes),
		FrequencyCount:  int(f.FrequencyCount),
		FrequencyType:   f.FrequencyType,
		LocationID:      f.LocationID,
		StartDate:       sharedmodels.DateOnly(parseDBTime(f.StartDate)),
		EndDate:         sharedmodels.DateOnly(parseDBTime(f.EndDate)),
		ProviderID:      f.ProviderID,
		CreatedAt:       parseDBTime(f.CreatedAt),
		UpdatedAt:       parseDBTime(f.UpdatedAt),
	}
}

type IEPState struct {
	created  bool
	archived bool
	deleted  bool
}

func (iep IEPState) isActive() bool {
	if iep.created && !iep.archived && !iep.deleted {
		return true
	}
	return false
}

type ServiceAddedToStudentEvent struct {
	ID      string       `json:"iep_service_added_to_student_event_id"`
	Service ServiceState `json:"service"`
	Scope   ServiceScope `json:"scope"`
}

type ServiceUpdatedEvent struct {
	ID      string       `json:"iep_service_updated_event_id"`
	Service ServiceState `json:"service"`
	Scope   ServiceScope `json:"scope"`
}

type ServiceArchivedEvent struct {
	EventID    string       `json:"iep_service_archived_event_id"`
	ServiceID  string       `json:"service.id"`
	ArchivedAt string       `json:"service.archived_at"`
	Scope      ServiceScope `json:"scope"`
}

type ServiceDeletedEvent struct {
	EventID   string       `json:"iep_service_deleted_event_id"`
	ServiceID string       `json:"service.id"`
	DeletedAt string       `json:"service.deleted_at"`
	Scope     ServiceScope `json:"scope"`
}

type ServiceScope struct {
	ServiceID string `json:"service_id"`
	IEPID     string `json:"iep_id"`
	StudentID string `json:"student_id"`
}

func NewServiceAddedToStudentEvent(
	cmd AddServiceToIEPCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	now := time.Now()
	cmd.Service.CreatedAt = now
	cmd.Service.UpdatedAt = now
	state := NewStateFromModel(cmd.Service)
	event := ServiceAddedToStudentEvent{
		ID:      cmd.Service.ID,
		Service: state,
		Scope:   serviceScope(cmd.Service.ID, cmd.Service.IEPID, cmd.Service.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   cmd.Service.ID,
		EventType: EventServiceAddedToIEP,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewServiceUpdatedEvent(
	eventID string,
	cmd UpdateServiceCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	now := time.Now()
	cmd.Service.UpdatedAt = now
	state := NewStateFromModel(cmd.Service)
	event := ServiceUpdatedEvent{
		ID:      eventID,
		Service: state,
		Scope:   serviceScope(cmd.Service.ID, cmd.Service.IEPID, cmd.Service.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewServiceDeletedEvent(
	eventID,
	serviceID,
	iepID,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ServiceDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     serviceScope(serviceID, iepID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func serviceScope(serviceID, iepID, studentID string) ServiceScope {
	return ServiceScope{
		ServiceID: serviceID,
		IEPID:     iepID,
		StudentID: studentID,
	}
}

func Channel(id string) string {
	return "services." + id
}

func ChannelAll() string {
	return "services.>"
}
