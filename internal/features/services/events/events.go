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
	StudentID       string       `json:"student_id"`
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
	StudentID       string       `json:"student_id"`
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
	StudentID string `json:"student_id"`
}

func NewServiceAddedToStudentEvent(
	eventID string,
	command AddServiceToIEPCommand,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ServiceAddedToStudentEvent{
		EventID:         eventID,
		ServiceID:       eventID,
		StudentID:       command.StudentID,
		ServiceName:     command.ServiceName,
		ServiceType:     command.ServiceType,
		IndirectMinutes: command.IndirectMinutes,
		DirectMinutes:   command.DirectMinutes,
		FrequencyCount:  command.FrequencyCount,
		FrequencyType:   command.FrequencyType,
		LocationID:      command.LocationID,
		StartDate:       command.StartDate,
		EndDate:         command.EndDate,
		ProviderID:      command.ProviderID,
		AddedAt:         addedAt.Format(time.RFC3339),
		Scope:           serviceScope(eventID, command.StudentID),
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
		StudentID:       studentID,
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
		Scope:           serviceScope(serviceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewServiceDeletedEvent(
	eventID string,
	serviceID string,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ServiceDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     serviceScope(serviceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func serviceScope(serviceID, studentID string) ServiceScope {
	return ServiceScope{
		ServiceID: serviceID,
		StudentID: studentID,
	}
}

func Channel(id string) string {
	return "iep_services." + id
}

func ChannelAll() string {
	return "iep_services.>"
}
