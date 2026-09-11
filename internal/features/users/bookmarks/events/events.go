package events

import (
	"time"

	"seek/internal/appdb"
	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type eventType = eventstore.EventType

const (
	EventStudentBookmarkAdded   eventType = "student_bookmark_added_event"
	EventStudentBookmarkRemoved eventType = "student_bookmark_removed_event"
)

type UserState struct {
	registered bool
	deleted    bool
}

type StudentState struct {
	created  bool
	archived bool
	deleted  bool
}

type StudentBookmarkAddedEvent struct {
	ID        string               `json:"student_bookmark_added_event_id"`
	UserID    string               `json:"user_id"`
	StudentID string               `json:"student_id"`
	AddedAt   string               `json:"added_at"`
	Scope     StudentBookmarkScope `json:"scope"`
}

type StudentBookmarkRemovedEvent struct {
	ID        string               `json:"student_bookmark_removed_event_id"`
	UserID    string               `json:"user_id"`
	StudentID string               `json:"student_id"`
	RemovedAt string               `json:"removed_at"`
	Scope     StudentBookmarkScope `json:"scope"`
}

type StudentBookmarkScope struct {
	UserID    string `json:"user_id"`
	StudentID string `json:"student_id"`
}

func NewStudentBookmarkAddedEvent(
	cmd AddStudentBookmarkCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	now := time.Now()
	id := uuidv7.NewString()
	event := StudentBookmarkAddedEvent{
		ID:        id,
		UserID:    cmd.UserID,
		StudentID: cmd.StudentID,
		AddedAt:   appdb.SQLTime(now),
		Scope:     studentBookmarkScope(cmd.UserID, cmd.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: EventStudentBookmarkAdded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentBookmarkRemovedEvent(
	cmd RemoveStudentBookmarkCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	id := uuidv7.NewString()
	event := StudentBookmarkRemovedEvent{
		ID:        id,
		UserID:    cmd.UserID,
		StudentID: cmd.StudentID,
		Scope:     studentBookmarkScope(cmd.UserID, cmd.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: EventStudentBookmarkRemoved,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func studentBookmarkScope(userID, studentID string) StudentBookmarkScope {
	return StudentBookmarkScope{
		UserID:    userID,
		StudentID: studentID,
	}
}
