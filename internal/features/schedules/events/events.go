package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	ScheduleCreated = "ScheduleCreated"
	ScheduleUpdated = "ScheduleUpdated"
	ScheduleDeleted = "ScheduleDeleted"
)

// EVENT FIELDS
const (
	ScheduleIDField        = "schedule_id"
	ScheduleTitleField     = "title"
	ScheduleTeacherIDField = "teacher_id"
	ScheduleCreatedIDField = "schedule_created_event_id"
	ScheduleCreatedAtField = "created_at"
	ScheduleUpdatedIDField = "schedule_updated_event_d"
	ScheduleUpdatedAtField = "updated_at"
	ScheduleDeletedIDField = "schedule_deleted_event_id"
	ScheduleDeletedAtField = "deleted_at"
	ScheduleScopeIDField   = "scope.schedule_id"
)

type ScheduleCreatedEvent struct {
	EventID   string        `json:"schedule_created_event_id"`
	Title     string        `json:"title"`
	TeacherID string        `json:"teacher_id"`
	CreatedAt string        `json:"created_at"`
	Scope     ScheduleScope `json:"scope"`
}

type ScheduleUpdatedEvent struct {
	EventID   string        `json:"schedule_updated_event_id"`
	Title     string        `json:"title"`
	TeacherID string        `json:"teacher_id"`
	UpdatedAt string        `json:"updated_at"`
	Scope     ScheduleScope `json:"scope"`
}

type ScheduleDeletedEvent struct {
	EventID   string        `json:"schedule_deleted_event_id"`
	DeletedAt string        `json:"deleted_at"`
	Scope     ScheduleScope `json:"scope"`
}

type ScheduleScope struct {
	ScheduleID string `json:"schedule_id"`
}

func NewScheduleCreatedEvent(
	scheduleID,
	title,
	teacherID string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ScheduleCreatedEvent{
		EventID:   scheduleID,
		Title:     title,
		TeacherID: teacherID,
		CreatedAt: createdAt.Format(time.RFC3339),
		Scope:     scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   scheduleID,
		EventType: ScheduleCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewScheduleUpdatedEvent(
	eventID,
	scheduleID,
	title,
	teacherID string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ScheduleUpdatedEvent{
		EventID:   eventID,
		Title:     title,
		TeacherID: teacherID,
		UpdatedAt: updatedAt.Format(time.RFC3339),
		Scope:     scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ScheduleUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewScheduleDeletedEvent(
	eventID,
	scheduleID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ScheduleDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ScheduleDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func scheduleScope(id string) ScheduleScope {
	return ScheduleScope{ScheduleID: id}
}

func Channel(id string) string {
	return "schedules." + id
}
