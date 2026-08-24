package events

import (
	"time"

	"seek/internal/eventstore"
)

type eventType = eventstore.EventType

// event types
const (
	EventIEPAddedToStudent     eventType = "iep_added_to_student_event"
	EventIEPRemovedFromStudent eventType = "iep_removed_from_student_event"
	EventIEPUpdated            eventType = "iep_updated_event"
	EventIEPArchived           eventType = "iep_archived_event"
	EventIEPDeleted            eventType = "iep_deleted_event"
)

// event fields for event IDs
const (
	FieldIEPAddedToStudentEventID     = "iep_added_to_student_event_id"
	FieldIEPRemovedFromStudentEventID = "iep_removed_from_student_event_id"
	FieldIEPUpdatedEventID            = "iep_updated_event_id"
	FieldIEPArchivedEventID           = "iep_archived_event_id"
	FieldIEPDeletedEventID            = "iep_deleted_event_id"
)

// event fields
const (
	FieldIEPID          = "iep_id"
	FieldIEPStudentID   = "student_id"
	FieldIEPStartDate   = "start_date"
	FieldIEPEndDate     = "end_date"
	FieldIEPAmendedDate = "amended_date"
	FieldIEPAddedAt     = "added_at"
	FieldIEPUpdatedAt   = "updated_at"
	FieldIEPArchivedAt  = "archived_at"
	FieldIEPDeletedAt   = "deleted_at"
	FieldIEPScopeID     = "scope.iep_id"
)

type IEPState struct {
	ID          string    `json:"iep_id"`
	StudentID   string    `json:"student_id"`
	StartDate   string    `json:"start_date,omitempty"`
	EndDate     string    `json:"end_date,omitempty"`
	AmendedDate string    `json:"amended_date,omitempty"`
	AddedAt     time.Time `json:"added_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	ArchivedAt  time.Time `json:"archived_at,omitempty"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
}

type IEPAddedToStudentEvent struct {
	EventID string   `json:"iep_added_to_student_event_id"`
	IEP     IEPState `json:"iep"`
	Scope   IEPScope `json:"scope"`
}

type IEPUpdatedEvent struct {
	EventID string   `json:"iep_updated_event_id"`
	IEP     IEPState `json:"iep"`
	Scope   IEPScope `json:"scope"`
}

type IEPArchivedEvent struct {
	EventID string   `json:"iep_archived_event_id"`
	IEP     IEPState `json:"iep"`
	Scope   IEPScope `json:"scope"`
}

type IEPDeletedEvent struct {
	EventID string   `json:"student_iep_deleted_event_id"`
	IEP     IEPState `json:"iep"`
	Scope   IEPScope `json:"scope"`
}

type IEPScope struct {
	IEPID     string `json:"iep_id"`
	StudentID string `json:"student_id"`
}

func NewStudentIEPAddedToStudentEvent(
	eventID string,
	command AddIEPToStudentCommand,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPAddedToStudentEvent{
		EventID: eventID,
		IEP:     command.IEP,
		Scope:   iepScope(eventID, command.IEP.StudentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPAddedToStudent,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentIEPUpdatedEvent(
	eventID string,
	command UpdateIEPCommand,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPUpdatedEvent{
		EventID: eventID,
		IEP:     command.IEP,
		Scope:   iepScope(command.IEP.ID, command.IEP.StudentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentIEPDeletedEvent(
	eventID string,
	IEPID string,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPDeletedEvent{
		EventID: eventID,
		Scope:   iepScope(IEPID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func iepScope(iepID, studentID string) IEPScope {
	return IEPScope{
		IEPID:     iepID,
		StudentID: studentID,
	}
}

func Channel(id string) string {
	return "student_ieps." + id
}

func ChannelAll() string {
	return "student_ieps.>"
}
