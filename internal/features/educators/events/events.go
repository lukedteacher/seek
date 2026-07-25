package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	EducatorCreated  = "EducatorCreated"
	EducatorUpdated  = "EducatorUpdated"
	EducatorArchived = "EducatorArchived"
	EducatorDeleted  = "EducatorDeleted"
)

// EVENT FIELDS
const (
	EducatorIDField              = "educator_id"
	EducatorGivenNameField       = "given_name"
	EducatorChosenNameField      = "chosen_name"
	EducatorFamilyNameField      = "family_name"
	EducatorEmailField           = "email"
	EducatorRoleField            = "role"
	EducatorCreatedEventIDField  = "educator_created_event_id"
	EducatorCreatedAtField       = "created_at"
	EducatorUpdatedEventIDField  = "educator_updated_event_id"
	EducatorUpdatedAtField       = "updated_at"
	EducatorArchivedEventIDField = "educator_archived_event_id"
	EducatorArchivedAtField      = "archived_at"
	EducatorDeletedEventIDField  = "educator_deleted_event_id"
	EducatorDeletedAtField       = "deleted_at"
	EducatorScopeIDField         = "scope.educator_created_event_id"
)

type EducatorCreatedEvent struct {
	EventID    string        `json:"educator_created_event_id"`
	GivenName  string        `json:"given_name"`
	ChosenName string        `json:"chosen_name"`
	FamilyName string        `json:"family_name"`
	Email      string        `json:"email"`
	Role       string        `json:"role"`
	CreatedAt  string        `json:"created_at"`
	Scope      EducatorScope `json:"scope"`
}

type EducatorUpdatedEvent struct {
	EventID    string        `json:"educator_created_event_id"`
	GivenName  string        `json:"given_name"`
	ChosenName string        `json:"chosen_name"`
	FamilyName string        `json:"family_name"`
	Email      string        `json:"email"`
	Role       string        `json:"role"`
	UpdatedAt  string        `json:"updated_at"`
	Scope      EducatorScope `json:"scope"`
}

type EducatorArchivedEvent struct {
	EventID    string        `json:"educator_created_event_id"`
	ArchivedAt string        `json:"archived_at"`
	Scope      EducatorScope `json:"scope"`
}

type EducatorDeletedEvent struct {
	EventID   string        `json:"educator_deleted_event_id"`
	DeletedAt string        `json:"deleted_at"`
	Scope     EducatorScope `json:"scope"`
}

type EducatorScope struct {
	ID string `json:"educator_created_event_id"`
}

func NewEducatorCreatedEvent(
	educatorID,
	givenName,
	chosenName,
	familyName,
	email,
	role string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorCreatedEvent{
		EventID:    educatorID,
		GivenName:  givenName,
		ChosenName: chosenName,
		FamilyName: familyName,
		Email:      email,
		Role:       role,
		CreatedAt:  createdAt.Format(time.RFC3339),
		Scope:      educatorScope(educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   educatorID,
		EventType: EducatorCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorUpdatedEvent(
	eventID,
	educatorID,
	givenName,
	chosenName,
	familyName,
	email,
	role string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorUpdatedEvent{
		EventID:    eventID,
		GivenName:  givenName,
		ChosenName: chosenName,
		FamilyName: familyName,
		Email:      email,
		Role:       role,
		UpdatedAt:  updatedAt.Format(time.RFC3339),
		Scope:      educatorScope(educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorArchivedEvent(
	eventID,
	educatorID string,
	archivedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorArchivedEvent{
		EventID:    eventID,
		ArchivedAt: archivedAt.Format(time.RFC3339),
		Scope:      educatorScope(educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorArchived,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorDeletedEvent(
	eventID,
	educatorID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     educatorScope(educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func educatorScope(id string) EducatorScope {
	return EducatorScope{ID: id}
}

func Channel(id string) string {
	return "educators." + id
}

func ChannelAll() string {
	return "educators.>"
}
