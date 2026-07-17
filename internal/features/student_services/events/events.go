package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	StudentServiceCreated = "StudentServiceCreated"
	StudentServiceUpdated = "StudentServiceUpdated"
	StudentServiceDeleted = "StudentServiceDeleted"
)

// EVENT METADATA
const (
	StudentServiceCreatedIDField = "student_service_created_event_id"
	StudentServiceUpdatedIDField = "student_service_updated_event_id"
	StudentServiceDeletedIDField = "student_service_deleted_event_id"
)

// EVENT FIELDS
const (
	StudentServiceIDField              = "service_id"
	StudentServiceStudentIDField       = "student_id"
	StudentServiceTypeField            = "service_type"
	StudentServiceIndirectMinutesField = "indirect_minutes"
	StudentServiceDirectMinutesField   = "direct_minutes"
	StudentServiceFrequencyCountField  = "frequency_count"
	StudentServiceFrequencyTypeField   = "frequency_type"
	StudentServiceLocationField        = "location"
	StudentServiceStartDateField       = "start_date"
	StudentServiceEndDateField         = "end_date"
	StudentServiceProviderField        = "provider"
	StudentServiceCreatedAtField       = "created_at"
	StudentServiceUpdatedAtField       = "updated_at"
	StudentServiceDeletedAtField       = "deleted_at"
	StudentServiceScopeIDField         = "scope.service_id"
)

// includes student_service scope which may be redundant
// since event ID is the same as the student_service for created
type StudentServiceCreatedEvent struct {
	EventID         string              `json:"student_service_created_event_id"`
	ServiceID       string              `json:"service_id"`
	StudentID       string              `json:"student_id"`
	ServiceType     string              `json:"service_type"`
	IndirectMinutes int                 `json:"indirect_minutes"`
	DirectMinutes   int                 `json:"direct_minutes"`
	FrequencyCount  int                 `json:"frequency_count"`
	FrequencyType   string              `json:"frequency_type"`
	Location        string              `json:"location"`
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	Provider        string              `json:"provider"`
	CreatedAt       string              `json:"created_at"`
	Scope           StudentServiceScope `json:"scope"`
}

type StudentServiceUpdatedEvent struct {
	EventID         string              `json:"student_service_updated_event_id"`
	ServiceID       string              `json:"service_id"`
	StudentID       string              `json:"student_id"`
	ServiceType     string              `json:"service_type"`
	IndirectMinutes int                 `json:"indirect_minutes"`
	DirectMinutes   int                 `json:"direct_minutes"`
	FrequencyCount  int                 `json:"frequency_count"`
	FrequencyType   string              `json:"frequency_type"`
	Location        string              `json:"location"`
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	Provider        string              `json:"provider"`
	UpdatedAt       string              `json:"updated_at"`
	Scope           StudentServiceScope `json:"scope"`
}

type StudentServiceDeletedEvent struct {
	EventID   string              `json:"student_service_deleted_event_id"`
	DeletedAt string              `json:"deleted_at"`
	Scope     StudentServiceScope `json:"scope"`
}

type StudentServiceScope struct {
	ServiceID string `json:"service_id"`
	StudentID string `json:"student_id"`
}

func NewStudentServiceCreatedEvent(
	serviceID string,
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
	event := StudentServiceCreatedEvent{
		EventID:         serviceID,
		ServiceID:       serviceID,
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
		Scope:           studentServiceScope(serviceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   serviceID,
		EventType: StudentServiceCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentServiceUpdatedEvent(
	eventID string,
	serviceID string,
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
	event := StudentServiceUpdatedEvent{
		EventID:         eventID,
		ServiceID:       serviceID,
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
		Scope:           studentServiceScope(serviceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentServiceUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentServiceDeletedEvent(
	eventID string,
	serviceID string,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentServiceDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     studentServiceScope(serviceID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentServiceDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func studentServiceScope(studentID, serviceID string) StudentServiceScope {
	return StudentServiceScope{
		StudentID: studentID,
		ServiceID: serviceID,
	}
}

func Channel(id string) string {
	return "student_service." + id
}

func ChannelAll() string {
	return "student_service.>"
}
