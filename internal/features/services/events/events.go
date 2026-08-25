package events

import (
	"time"

	"seek/internal/eventstore"
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

type ServiceAddedToStudentEvent struct {
	EventID         string       `json:"iep_service_added_to_student_event_id"`
	ServiceID       string       `json:"iep_service_id"`
	IEPID           string       `json:"iep_id"`
	ServiceName     string       `json:"service_name"`
	ServiceType     string       `json:"service_type"`
	IndirectMinutes int          `json:"indirect_minutes"`
	DirectMinutes   int          `json:"direct_minutes"`
	FrequencyCount  int          `json:"frequency_count"`
	FrequencyType   string       `json:"frequency_type"`
	LocationID      string       `json:"location_id"`
	StartDate       string       `json:"start_date"`
	EndDate         string       `json:"end_date"`
	ProviderID      string       `json:"provider_id"`
	AddedAt         string       `json:"added_at"`
	Scope           ServiceScope `json:"scope"`
}

type ServiceUpdatedEvent struct {
	EventID         string       `json:"iep_service_updated_event_id"`
	ServiceID       string       `json:"iep_service_id"`
	IEPID           string       `json:"iep_id"`
	ServiceName     string       `json:"service_name"`
	ServiceType     string       `json:"service_type"`
	IndirectMinutes int          `json:"indirect_minutes"`
	DirectMinutes   int          `json:"direct_minutes"`
	FrequencyCount  int          `json:"frequency_count"`
	FrequencyType   string       `json:"frequency_type"`
	LocationID      string       `json:"location_id"`
	StartDate       string       `json:"start_date"`
	EndDate         string       `json:"end_date"`
	Provider        string       `json:"provider"`
	UpdatedAt       string       `json:"updated_at"`
	Scope           ServiceScope `json:"scope"`
}

type ServiceDeletedEvent struct {
	EventID   string       `json:"iep_service_deleted_event_id"`
	DeletedAt string       `json:"deleted_at"`
	Scope     ServiceScope `json:"scope"`
}

type ServiceScope struct {
	ServiceID string `json:"iep_service_added_to_student_event_id"`
	IEPID     string `json:"iep_id"`
	StudentID string `json:"student_id"`
}

func NewServiceAddedToStudentEvent(
	eventID string,
	cmd AddServiceToIEPCommand,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ServiceAddedToStudentEvent{
		EventID:         eventID,
		ServiceID:       eventID,
		IEPID:           cmd.IEPID,
		ServiceName:     cmd.ServiceName,
		ServiceType:     cmd.ServiceType,
		IndirectMinutes: cmd.IndirectMinutes,
		DirectMinutes:   cmd.DirectMinutes,
		FrequencyCount:  cmd.FrequencyCount,
		FrequencyType:   cmd.FrequencyType,
		LocationID:      cmd.LocationID,
		StartDate:       cmd.StartDate,
		EndDate:         cmd.EndDate,
		ProviderID:      cmd.ProviderID,
		AddedAt:         addedAt.Format(time.RFC3339),
		Scope:           serviceScope(eventID, cmd.IEPID, cmd.StudentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceAddedToIEP,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewServiceUpdatedEvent(
	eventID,
	serviceID,
	iepID,
	studentID,
	serviceName,
	serviceType string,
	indirectMinutes,
	directMinutes,
	frequencyCount int,
	frequencyType,
	location,
	startDate,
	endDate,
	provider string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ServiceUpdatedEvent{
		EventID:         eventID,
		ServiceID:       serviceID,
		IEPID:           iepID,
		ServiceName:     serviceName,
		ServiceType:     serviceType,
		IndirectMinutes: indirectMinutes,
		DirectMinutes:   directMinutes,
		FrequencyCount:  frequencyCount,
		FrequencyType:   frequencyType,
		LocationID:      location,
		StartDate:       startDate,
		EndDate:         endDate,
		Provider:        provider,
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		Scope:           serviceScope(serviceID, iepID, studentID),
	}
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
	return "iep_services." + id
}

func ChannelAll() string {
	return "iep_services.>"
}
