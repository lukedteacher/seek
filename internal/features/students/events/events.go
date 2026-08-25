package events

import (
	"time"

	"seek/internal/eventstore"
)

type eventType = eventstore.EventType

// student event types
const (
	EventStudentCreated  eventType = "student_created_event"
	EventStudentUpdated  eventType = "student_updated_event"
	EventStudentArchived eventType = "student_archived_event"
	EventStudentDeleted  eventType = "student_deleted_event"
)

// student event id & scope fields
const (
	FieldStudentCreatedEventID  = "student_created_event_id"
	FieldStudentUpdatedEventID  = "student_updated_event_id"
	FieldStudentArchivedEventID = "student_archived_event_id"
	FieldStudentDeletedEventID  = "student_deleted_event_id"
)

// student event fields
const (
	FieldStudentID         = "student_id"
	FieldStudentMARSSID    = "marss_id"
	FieldStudentGivenName  = "given_name"
	FieldStudentChosenName = "chosen_name"
	FieldStudentFamilyName = "family_name"
	FieldStudentEmail      = "email"
	FieldStudentUsername   = "username"
	FieldStudentGrade      = "grade"
	FieldStudentHomeroomID = "homeroom_id"
	FieldStudentPlanType   = "plan_type"
	FieldStudentCreatedAt  = "created_at"
	FieldStudentUpdatedAt  = "updated_at"
	FieldStudentArchivedAt = "archived_at"
	FieldStudentDeletedAt  = "deleted_at"
	FieldScopeStudentID    = "scope.student_id"
)

type StudentCreatedEvent struct {
	EventID string `json:"student_created_event_id"`
	StudentState
	Scope StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	EventID string `json:"student_updated_event_id"`
	StudentState
	Scope StudentScope `json:"scope"`
}

type StudentArchivedEvent struct {
	EventID    string       `json:"student_archived_event_id"`
	ArchivedAt string       `json:"archived_at"`
	Scope      StudentScope `json:"scope"`
}

type StudentDeletedEvent struct {
	EventID   string       `json:"student_deleted_event_id"`
	DeletedAt string       `json:"deleted_at"`
	Scope     StudentScope `json:"scope"`
}

type StudentScope struct {
	StudentID string `json:"student_id"`
}

func NewStudentCreatedEvent(
	student StudentState,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	student.CreatedAt = createdAt
	student.UpdatedAt = createdAt
	event := StudentCreatedEvent{
		EventID:      student.ID,
		StudentState: student,
		Scope:        studentScope(student.ID),
	}
	return eventstore.DomainEvent{
		EventID:   student.ID,
		EventType: EventStudentCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentUpdatedEvent(
	eventID string,
	student StudentState,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	student.UpdatedAt = updatedAt
	event := StudentUpdatedEvent{
		EventID:      eventID,
		StudentState: student,
		Scope:        studentScope(student.ID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventStudentUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentArchivedEvent(
	eventID,
	studentID string,
	archivedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentArchivedEvent{
		EventID:    eventID,
		ArchivedAt: archivedAt.Format(time.RFC3339),
		Scope:      studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventStudentArchived,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentDeletedEvent(
	eventID,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventStudentDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func studentScope(id string) StudentScope {
	return StudentScope{StudentID: id}
}

func Channel(id string) string {
	return "students." + id
}

func ChannelAll() string {
	return "students.>"
}
