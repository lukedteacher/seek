package events

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/homerooms/models"
	"seek/pkg/uuidv7"
)

type eventType = eventstore.EventType

// event types
const (
	EventHomeroomCreated             eventType = "homeroom_created_event"
	EventHomeroomUpdated             eventType = "homeroom_updated_event"
	EventHomeroomArchived            eventType = "homeroom_archived_event"
	EventHomeroomDeleted             eventType = "homeroom_deleted_event"
	EventEducatorAddedToHomeroom     eventType = "educator_added_to_homeroom_event"
	EventEducatorRemovedFromHomeroom eventType = "educator_removed_from_homeroom_event"
	EventStudentAddedToHomeroom      eventType = "student_added_to_homeroom_event"
	EventStudentRemovedFromHomeroom  eventType = "student_removed_from_homeroom_event"
)

// event ID fields
const (
	FieldHomeroomCreatedEventID             = "homeroom_created_event_id"
	FieldHomeroomUpdatedEventID             = "homeroom_updated_event_id"
	FieldHomeroomArchivedEventID            = "homeroom_archived_event_id"
	FieldHomeroomDeletedEventID             = "homeroom_deleted_event_id"
	FieldEducatorAddedToHomeroomEventID     = "educator_added_to_homeroom_event_id"
	EventEducatorRemovedFromHomeroomEventID = "educator_removed_from_homeroom_event_id"
	EventStudentAddedToHomeroomEventID      = "student_added_to_homeroom_event_id"
	EventStudentRemovedFromHomeroomEventID  = "student_removed_from_homeroom_event_id"
)

// event data fields
const (
	FieldHomeroomID                 = "homeroom_id"
	FieldHomeroomTitle              = "title"
	FieldHomeroomLocationID         = "location_id"
	FieldHomeroomCreatedAt          = "created_at"
	FieldHomeroomUpdatedAt          = "updated_at"
	FieldHomeroomArchivedAt         = "archived_at"
	FieldHomeroomDeletedAt          = "deleted_at"
	FieldHomeroomEducatorEducatorID = "educator_id"
	FieldHomeroomEducatorAddedAt    = "added_at"
	FieldHomeroomEducatorRemovedAt  = "removed_at"
	FieldHomeroomStudentStudentID   = "student_id"
	FieldHomeroomStudentAddedAt     = "added_at"
	FieldHomeroomStudentRemovedAt   = "removed_at"
)

// event scope fields
const (
	FieldHomeroomScopeID                 = "scope.homeroom_id"
	FieldHomeroomEducatorScopeEducatorID = "scope.educator_id"
	FieldHomeroomStudentScopeStudentID   = "scope.student_id"
)

type HomeroomCreatedEvent struct {
	ID            string        `json:"homeroom_created_event_id"`
	Title         string        `json:"title"`
	GradesBitmask int           `json:"grades_bitmask"`
	LocationID    string        `json:"location_id"`
	Image         string        `json:"image"`
	CreatedAt     string        `json:"created_at"`
	Scope         HomeroomScope `json:"scope"`
}

type HomeroomUpdatedEvent struct {
	ID            string        `json:"homeroom_updated_event_id"`
	Title         string        `json:"title"`
	GradesBitmask int           `json:"grades_bitmask"`
	LocationID    string        `json:"location_id"`
	Image         string        `json:"image"`
	UpdatedAt     string        `json:"updated_at"`
	Scope         HomeroomScope `json:"scope"`
}

type HomeroomArchivedEvent struct {
	ID         string        `json:"homeroom_archived_event_id"`
	ArchivedAt string        `json:"archived_at"`
	Scope      HomeroomScope `json:"scope"`
}

type HomeroomDeletedEvent struct {
	ID        string        `json:"homeroom_deleted_event_id"`
	DeletedAt string        `json:"deleted_at"`
	Scope     HomeroomScope `json:"scope"`
}

type HomeroomScope struct {
	ID string `json:"homeroom_id"`
}

type EducatorAddedToHomeroomEvent struct {
	ID         string                `json:"educator_added_to_homeroom_event_id"`
	HomeroomID string                `json:"homeroom_id"`
	EducatorID string                `json:"educator_id"`
	AddedAt    time.Time             `json:"added_at"`
	Scope      HomeroomEducatorScope `json:"scope"`
}

type EducatorRemovedFromHomeroomEvent struct {
	ID         string                `json:"educator_removed_from_homeroom_event_id"`
	HomeroomID string                `json:"homeroom_id"`
	EducatorID string                `json:"educator_id"`
	RemovedAt  time.Time             `json:"removed_at"`
	Scope      HomeroomEducatorScope `json:"scope"`
}

type HomeroomEducatorScope struct {
	HomeroomID string `json:"homeroom_id"`
	EducatorID string `json:"educator_id"`
}

type StudentAddedToHomeroomEvent struct {
	ID         string               `json:"student_added_to_homeroom_event_id"`
	HomeroomID string               `json:"homeroom_id"`
	StudentID  string               `json:"student_id"`
	AddedAt    time.Time            `json:"added_at"`
	Scope      HomeroomStudentScope `json:"scope"`
}

type StudentRemovedFromHomeroomEvent struct {
	ID         string               `json:"student_removed_from_homeroom_event_id"`
	HomeroomID string               `json:"homeroom_id"`
	StudentID  string               `json:"student_id"`
	RemovedAt  time.Time            `json:"removed_at"`
	Scope      HomeroomStudentScope `json:"scope"`
}

type HomeroomStudentScope struct {
	HomeroomID string `json:"homeroom_id"`
	StudentID  string `json:"student_id"`
}

func NewHomeroomCreatedEvent(
	eventID string,
	homeroom models.Homeroom,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := HomeroomCreatedEvent{
		ID:            eventID,
		Title:         homeroom.Title,
		GradesBitmask: int(homeroom.GradesBitmask),
		LocationID:    homeroom.LocationID,
		Image:         homeroom.Image,
		CreatedAt:     createdAt.Format(time.RFC3339),
		Scope:         homeroomScope(eventID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventHomeroomCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewHomeroomUpdatedEvent(
	homeroom models.Homeroom,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := HomeroomUpdatedEvent{
		ID:            eventID,
		Title:         homeroom.Title,
		GradesBitmask: int(homeroom.GradesBitmask),
		LocationID:    homeroom.LocationID,
		Image:         homeroom.Image,
		UpdatedAt:     updatedAt.Format(time.RFC3339),
		Scope:         homeroomScope(homeroom.ID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventHomeroomUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewHomeroomArchivedEvent(
	homeroomID string,
	archivedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := HomeroomArchivedEvent{
		ID:         eventID,
		ArchivedAt: archivedAt.Format(time.RFC3339),
		Scope:      homeroomScope(homeroomID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventHomeroomArchived,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewHomeroomDeletedEvent(
	homeroomID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := HomeroomDeletedEvent{
		ID:        eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     homeroomScope(homeroomID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventHomeroomDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorAddedToHomeroomEvent(
	homeroomID,
	educatorID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := EducatorAddedToHomeroomEvent{
		ID:         eventID,
		HomeroomID: homeroomID,
		EducatorID: educatorID,
		AddedAt:    addedAt,
		Scope:      homeroomEducatorScope(homeroomID, educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorAddedToHomeroom,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorRemovedFromHomeroomEvent(
	homeroomID,
	educatorID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := EducatorRemovedFromHomeroomEvent{
		ID:         eventID,
		HomeroomID: homeroomID,
		EducatorID: educatorID,
		RemovedAt:  removedAt,
		Scope:      homeroomEducatorScope(homeroomID, educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorRemovedFromHomeroom,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentAddedToHomeroomEvent(
	homeroomID,
	studentID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := StudentAddedToHomeroomEvent{
		ID:         eventID,
		HomeroomID: homeroomID,
		StudentID:  studentID,
		AddedAt:    addedAt,
		Scope:      homeroomStudentScope(homeroomID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventStudentAddedToHomeroom,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentRemovedFromHomeroomEvent(
	homeroomID,
	studentID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	eventID := uuidv7.NewString()
	event := StudentRemovedFromHomeroomEvent{
		ID:         eventID,
		HomeroomID: homeroomID,
		StudentID:  studentID,
		RemovedAt:  removedAt,
		Scope:      homeroomStudentScope(homeroomID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventStudentRemovedFromHomeroom,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func homeroomScope(id string) HomeroomScope {
	return HomeroomScope{ID: id}
}

func homeroomEducatorScope(homeroomID, educatorID string) HomeroomEducatorScope {
	return HomeroomEducatorScope{
		HomeroomID: homeroomID,
		EducatorID: educatorID,
	}
}

func homeroomStudentScope(homeroomID, studentID string) HomeroomStudentScope {
	return HomeroomStudentScope{
		HomeroomID: homeroomID,
		StudentID:  studentID,
	}
}

func Channel(id string) string {
	return "homerooms." + id
}

func ChannelAll() string {
	return "homerooms.>"
}
