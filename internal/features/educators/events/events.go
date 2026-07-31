package events

import (
	"time"
)

// educator event types
const (
	EducatorCreated  = "EducatorCreated"
	EducatorUpdated  = "EducatorUpdated"
	EducatorArchived = "EducatorArchived"
	EducatorDeleted  = "EducatorDeleted"
)

// educator event id and scope fields
const (
	FieldEducatorCreatedEventID  = "educator_created_event_id"
	FieldEducatorUpdatedEventID  = "educator_updated_event_id"
	FieldEducatorArchivedEventID = "educator_archived_event_id"
	FieldEducatorDeletedEventID  = "educator_deleted_event_id"
	FieldEducatorScopeID         = "scope.educator_created_event_id"
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
	GivenName  string `json:"given_name"`
	ChosenName string `json:"chosen_name"`
	FamilyName string `json:"family_name"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Role       string `json:"role"`
}

type EducatorCreatedEvent struct {
	EventID string `json:"educator_created_event_id"`
	EducatorState
	CreatedAt time.Time     `json:"created_at"`
	Scope     EducatorScope `json:"scope"`
}

type EducatorUpdatedEvent struct {
	EventID string `json:"educator_updated_event_id"`
	EducatorState
	UpdatedAt time.Time     `json:"updated_at"`
	Scope     EducatorScope `json:"scope"`
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
	ID string `json:"educator_created_event_id"`
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
