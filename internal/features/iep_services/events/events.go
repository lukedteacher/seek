package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	IEPServiceCreated = "IEPServiceCreated"
	IEPServiceUpdated = "IEPServiceUpdated"
	IEPServiceDeleted = "IEPServiceDeleted"
)

// EVENT METADATA
const (
	IEPServiceCreatedIDField = "iep_service_created_event_id"
	IEPServiceUpdatedIDField = "iep_service_updated_event_id"
	IEPServiceDeletedIDField = "iep_service_deleted_event_id"
)

// EVENT FIELDS
const (
	IEPServiceIDField              = "iep_service_id"
	IEPServiceStudentIDField       = "student_id"
	IEPServiceServiceTypeField     = "service_type"
	IEPServiceIndirectMinutesField = "indirect_minutes"
	IEPServiceDirectMinutesField   = "direct_minutes"
	IEPServiceFrequencyCountField  = "frequency_count"
	IEPServiceFrequencyTypeField   = "frequency_type"
	IEPServiceLocationField        = "location"
	IEPServiceStartDateField       = "start_date"
	IEPServiceEndDateField         = "end_date"
	IEPServiceProviderField        = "provider"
	IEPServiceCreatedAtField       = "created_at"
	IEPServiceUpdatedAtField       = "updated_at"
	IEPServiceDeletedAtField       = "deleted_at"
	IEPServiceScopeIDField         = "scope.iep_service_id"
)

// includes iep_service scope which may be redundant
// since event ID is the same as the iep_service for created
type IEPServiceCreatedEvent struct {
	EventID         string          `json:"iep_service_created_event_id"`
	IEPServiceID    string          `json:"iep_service_id"`
	StudentID       string          `json:"student_id"`
	ServiceType     string          `json:"service_type"`
	IndirectMinutes int             `json:"indirect_minutes"`
	DirectMinutes   int             `json:"direct_minutes"`
	FrequencyCount  int             `json:"frequency_count"`
	FrequencyType   string          `json:"frequency_type"`
	Location        string          `json:"location"`
	StartDate       string          `json:"start_date"`
	EndDate         string          `json:"end_date"`
	Provider        string          `json:"provider"`
	CreatedAt       string          `json:"created_at"`
	Scope           IEPServiceScope `json:"scope"`
}

type IEPServiceUpdatedEvent struct {
	EventID         string          `json:"iep_service_updated_event_id"`
	IEPServiceID    string          `json:"iep_service_id"`
	StudentID       string          `json:"student_id"`
	ServiceType     string          `json:"service_type"`
	IndirectMinutes int             `json:"indirect_minutes"`
	DirectMinutes   int             `json:"direct_minutes"`
	FrequencyCount  int             `json:"frequency_count"`
	FrequencyType   string          `json:"frequency_type"`
	Location        string          `json:"location"`
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
	IEPServiceID string `json:"iep_service_id"`
	StudentID    string `json:"student_id"`
}

func NewIEPServiceCreatedEvent(
	iepServiceID string,
	studentID string,
	serviceType string,
	indirectMinutes int,
	directMinutes int,
	frequencyCount int,
	frequencyType string,
	location string,
	startDate string,
	endDate string,
	provider string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPServiceCreatedEvent{
		EventID:         iepServiceID,
		IEPServiceID:    iepServiceID,
		StudentID:       studentID,
		ServiceType:     serviceType,
		IndirectMinutes: indirectMinutes,
		DirectMinutes:   directMinutes,
		FrequencyCount:  frequencyCount,
		FrequencyType:   frequencyType,
		Location:        location,
		StartDate:       startDate,
		EndDate:         endDate,
		Provider:        provider,
		CreatedAt:       createdAt.Format(time.RFC3339),
		Scope:           iepServiceScope(iepServiceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   iepServiceID,
		EventType: IEPServiceCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewIEPServiceUpdatedEvent(
	eventID string,
	iepServiceID string,
	studentID string,
	serviceType string,
	indirectMinutes int,
	directMinutes int,
	frequencyCount int,
	frequencyType string,
	location string,
	startDate string,
	endDate string,
	provider string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPServiceUpdatedEvent{
		EventID:         eventID,
		IEPServiceID:    iepServiceID,
		StudentID:       studentID,
		ServiceType:     serviceType,
		IndirectMinutes: indirectMinutes,
		DirectMinutes:   directMinutes,
		FrequencyCount:  frequencyCount,
		FrequencyType:   frequencyType,
		Location:        location,
		StartDate:       startDate,
		EndDate:         endDate,
		Provider:        provider,
		UpdatedAt:       updatedAt.Format(time.RFC3339),
		Scope:           iepServiceScope(iepServiceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: IEPServiceUpdated,
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
		EventType: IEPServiceDeleted,
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
	return "iep_service." + id
}

func ChannelAll() string {
	return "iep_service.>"
}
