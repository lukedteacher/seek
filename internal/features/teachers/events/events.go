package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	TeacherCreated = "TeacherCreated"
	TeacherUpdated = "TeacherUpdated"
	TeacherDeleted = "TeacherDeleted"
)

// EVENT FIELDS
const (
	TeacherIDField         = "teacher_id"
	TeacherGivenNameField  = "given_name"
	TeacherChosenNameField = "chosen_name"
	TeacherFamilyNameField = "family_name"
	TeacherCreatedIDField  = "teacher_created_event_id"
	TeacherCreatedAtField  = "created_at"
	TeacherUpdatedIDField  = "teacher_updated_event_id"
	TeacherUpdatedAtField  = "updated_at"
	TeacherDeletedIDField  = "teacher_deleted_event_id"
	TeacherDeletedAtField  = "deleted_at"
	TeacherScopeIDField    = "scope.teacher_id"
)

type TeacherCreatedEvent struct {
	EventID    string       `json:"teacher_created_event_id"`
	GivenName  string       `json:"given_name"`
	ChosenName string       `json:"chosen_name"`
	FamilyName string       `json:"family_name"`
	CreatedAt  string       `json:"created_at"`
	Scope      TeacherScope `json:"scope"`
}

type TeacherUpdatedEvent struct {
	EventID    string       `json:"teacher_updated_event_id"`
	GivenName  string       `json:"given_name"`
	ChosenName string       `json:"chosen_name"`
	FamilyName string       `json:"family_name"`
	UpdatedAt  string       `json:"updated_at"`
	Scope      TeacherScope `json:"scope"`
}

type TeacherDeletedEvent struct {
	EventID   string       `json:"teacher_deleted_event_id"`
	DeletedAt string       `json:"deleted_at"`
	Scope     TeacherScope `json:"scope"`
}

type TeacherScope struct {
	ID string `json:"teacher_id"`
}

func NewTeacherCreatedEvent(
	teacherID,
	givenName,
	chosenName,
	familyName string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := TeacherCreatedEvent{
		EventID:    teacherID,
		GivenName:  givenName,
		ChosenName: chosenName,
		FamilyName: familyName,
		CreatedAt:  createdAt.Format(time.RFC3339),
		Scope:      teacherScope(teacherID),
	}
	return eventstore.DomainEvent{
		EventID:   teacherID,
		EventType: TeacherCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewTeacherUpdatedEvent(
	eventID,
	teacherID,
	givenName,
	chosenName,
	familyName string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := TeacherUpdatedEvent{
		EventID:    eventID,
		GivenName:  givenName,
		ChosenName: chosenName,
		FamilyName: familyName,
		UpdatedAt:  updatedAt.Format(time.RFC3339),
		Scope:      teacherScope(teacherID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: TeacherUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewTeacherDeletedEvent(
	eventID,
	teacherID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := TeacherDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     teacherScope(teacherID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: TeacherDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func teacherScope(id string) TeacherScope {
	return TeacherScope{ID: id}
}

func Channel(id string) string {
	return "teachers." + id
}
