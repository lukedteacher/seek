package events

import (
	"time"

	"seek/internal/eventstore"
)

type eventType = eventstore.EventType

// educator event types
const (
	EventEducatorCreated  eventType = "educator_created"
	EventEducatorUpdated  eventType = "educator_updated"
	EventEducatorArchived eventType = "educator_archived"
	EventEducatorDeleted  eventType = "educator_deleted"
)

// educator event id and scope fields
const (
	FieldEducatorCreatedEventID  = "educator_created_event_id"
	FieldEducatorUpdatedEventID  = "educator_updated_event_id"
	FieldEducatorArchivedEventID = "educator_archived_event_id"
	FieldEducatorDeletedEventID  = "educator_deleted_event_id"
	FieldScopeEducatorID         = "scope.educator_id"
)

// educator event fields
const (
	FieldEducatorID         = "educator_id"
	FieldEducatorGivenName  = "given_name"
	FieldEducatorChosenName = "chosen_name"
	FieldEducatorFamilyName = "family_name"
	FieldEducatorEmail      = "email"
	FieldEducatorUsername   = "username"
	FieldEducatorRole       = "role"
	FieldEducatorCreatedAt  = "created_at"
	FieldEducatorUpdatedAt  = "updated_at"
	FieldEducatorArchivedAt = "archived_at"
	FieldEducatorDeletedAt  = "deleted_at"
)

type EducatorState struct {
	ID         string    `json:"id"`
	GivenName  string    `json:"given_name"`
	ChosenName string    `json:"chosen_name"`
	FamilyName string    `json:"family_name"`
	Pronouns   string    `json:"pronouns"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	Roles      []string  `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EducatorCreatedEvent struct {
	EventID string `json:"educator_created_event_id"`
	EducatorState
	Scope EducatorScope `json:"scope"`
}

type EducatorUpdatedEvent struct {
	EventID string `json:"educator_updated_event_id"`
	EducatorState
	Scope EducatorScope `json:"scope"`
}

type EducatorArchivedEvent struct {
	EventID    string        `json:"educator_archived_event_id"`
	ArchivedAt time.Time     `json:"archived_at"`
	Scope      EducatorScope `json:"scope"`
}

type EducatorDeletedEvent struct {
	EventID   string        `json:"educator_deleted_event_id"`
	DeletedAt time.Time     `json:"deleted_at"`
	Scope     EducatorScope `json:"scope"`
}

type EducatorScope struct {
	ID string `json:"educator_id"`
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
