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
	EventIEPServiceUpdated     eventType = "iep_service_updated_event"
	EventIEPServiceArchived    eventType = "iep_service_archived_event"
	EventIEPServiceDeleted     eventType = "iep_service_deleted_event"
)

// event fields for event IDs
const (
	FieldServiceAddedToIEPEventID = "iep_service_added_to_student_event_id"
	FieldIEPServiceUpdatedEventID = "iep_service_updated_event_id"
	FieldIEPServiceDeletedEventID = "iep_service_deleted_event_id"
)

// event fields
const (
	FieldIEPServiceID              = "iep_service_id"
	FieldIEPServiceIEPID           = "iep_id"
	FieldIEPServiceStudentID       = "student_id"
	FieldIEPServiceServiceName     = "service_name"
	FieldIEPServiceServiceType     = "service_type"
	FieldIEPServiceIndirectMinutes = "indirect_minutes"
	FieldIEPServiceDirectMinutes   = "direct_minutes"
	FieldIEPServiceFrequencyCount  = "frequency_count"
	FieldIEPServiceFrequencyType   = "frequency_type"
	FieldIEPServiceLocationID      = "location_id"
	FieldIEPServiceStartDate       = "start_date"
	FieldIEPServiceEndDate         = "end_date"
	FieldIEPServiceProviderID      = "provider_id"
	FieldIEPServiceAddedAt         = "added_at"
	FieldIEPServiceUpdatedAt       = "updated_at"
	FieldIEPServiceArchivedAt      = "archived_at"
	FieldIEPServiceDeletedAt       = "deleted_at"
	FieldIEPServiceScopeID         = "scope.iep_service_added_to_student_event_id"
)

type IEPServiceAddedToStudentEvent struct {
	EventID         string          `json:"iep_service_added_to_student_event_id"`
	IEPServiceID    string          `json:"iep_service_id"`
	StudentID       string          `json:"student_id"`
	ServiceName     string          `json:"service_name"`
	ServiceType     string          `json:"service_type"`
	IndirectMinutes int             `json:"indirect_minutes"`
	DirectMinutes   int             `json:"direct_minutes"`
	FrequencyCount  int             `json:"frequency_count"`
	FrequencyType   string          `json:"frequency_type"`
	LocationID      string          `json:"location_id"`
	StartDate       string          `json:"start_date"`
	EndDate         string          `json:"end_date"`
	ProviderID      string          `json:"provider_id"`
	AddedAt         string          `json:"added_at"`
	Scope           IEPServiceScope `json:"scope"`
}

type IEPServiceUpdatedEvent struct {
	EventID         string          `json:"iep_service_updated_event_id"`
	IEPServiceID    string          `json:"iep_service_id"`
	StudentID       string          `json:"student_id"`
	ServiceName     string          `json:"service_name"`
	ServiceType     string          `json:"service_type"`
	IndirectMinutes int             `json:"indirect_minutes"`
	DirectMinutes   int             `json:"direct_minutes"`
	FrequencyCount  int             `json:"frequency_count"`
	FrequencyType   string          `json:"frequency_type"`
	LocationID      string          `json:"location_id"`
	StartDate       string          `json:"start_date"`
	EndDate         string          `json:"end_date"`
	Provider        string          `json:"provider"`
	UpdatedAt       string          `json:"updated_at"`
	Scope           IEPServiceScope `json:"scope"`
}

type IEPServiceDeletedEvent struct {
	EventID   string          `json:"iep_service_deleted_event_id"`
	DeletedAt string          `json:"deleted_at"`
	Scope     IEPServiceScope `json:"scope"`
}

type IEPServiceScope struct {
	IEPServiceID string `json:"iep_service_added_to_student_event_id"`
	StudentID    string `json:"student_id"`
}

func NewIEPServiceAddedToStudentEvent(
	eventID string,
	command AddServiceToIEPCommand,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPServiceAddedToStudentEvent{
		EventID:         eventID,
		IEPServiceID:    eventID,
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
		Scope:           iepServiceScope(eventID, command.StudentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventServiceAddedToIEP,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewIEPServiceUpdatedEvent(
	eventID,
	iepServiceID,
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
	event := IEPServiceUpdatedEvent{
		EventID:         eventID,
		IEPServiceID:    iepServiceID,
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
		Scope:           iepServiceScope(iepServiceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPServiceUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewIEPServiceDeletedEvent(
	eventID string,
	iepServiceID string,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPServiceDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     iepServiceScope(iepServiceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPServiceDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func iepServiceScope(iepServiceID, studentID string) IEPServiceScope {
	return IEPServiceScope{
		IEPServiceID: iepServiceID,
		StudentID:    studentID,
	}
}

func Channel(id string) string {
	return "iep_services." + id
}

func ChannelAll() string {
	return "iep_services.>"
}
