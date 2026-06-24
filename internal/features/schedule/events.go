package schedule

import (
	"time"

	"seek/internal/eventstore"
)

const (
	ScheduleCreated = "ScheduleCreated"
	ScheduleUpdated = "ScheduleUpdated"
	ScheduleDeleted = "ScheduleDeleted"
)

const (
	ScheduleIDField               = "scheduleId"
	ScheduleUserRegisteredIDField = "userRegisteredId"
	ScheduleTitleField            = "title"
	ScheduleCreatedAtField        = "createdAt"
	ScheduleRenamedIDField        = "scheduleRenamedId"
	ScheduleRenamedAtField        = "renamedAt"
	ScheduleDeletedIDField        = "scheduleDeletedId"
	ScheduleDeletedAtField        = "deletedAt"
	ScheduleScopeIDField          = "scope.id"
)

type ScheduleCreatedEvent struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	TeacherId string        `json:"teacher_id"`
	Duration  int64         `json:"duration"`
	Days      int64         `json:"days"`
	CreatedAt string        `json:"createdAt"`
	Scope     ScheduleScope `json:"scope"`
}

type ScheduleUpdatedEvent struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	TeacherId string        `json:"teacher_id"`
	Duration  int64         `json:"duration"`
	Days      int64         `json:"days"`
	UpdatedAt string        `json:"updatedAt"`
	Scope     ScheduleScope `json:"scope"`
}

type ScheduleDeletedEvent struct {
	ScheduleDeletedID string        `json:"scheduleDeletedId"`
	DeletedAt         string        `json:"deletedAt"`
	Scope             ScheduleScope `json:"scope"`
}

type ScheduleScope struct {
	ID               string `json:"id"`
	UserRegisteredID string `json:"userRegisteredId"`
}

func NewScheduleCreatedEvent(id, title, teacherId string, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: ScheduleCreated,
		Data: eventstore.MustData(ScheduleCreatedEvent{
			ID:        id,
			Title:     title,
			TeacherId: teacherId,
			CreatedAt: createdAt.Format(time.RFC3339),
			Scope:     scheduleScope(id),
		}),
		Metadata: metadata,
	}
}

func NewScheduleUpdatedEvent(eventID, scheduleID, title, teacherId string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ScheduleUpdated,
		Data: eventstore.MustData(ScheduleUpdatedEvent{
			ID:        eventID,
			Title:     title,
			TeacherId: teacherId,
			UpdatedAt: updatedAt.Format(time.RFC3339),
			Scope:     scheduleScope(scheduleID),
		}),
		Metadata: metadata,
	}
}

func NewScheduleDeletedEvent(scheduleDeletedID, id string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   scheduleDeletedID,
		EventType: ScheduleDeleted,
		Data: eventstore.MustData(ScheduleDeletedEvent{
			ScheduleDeletedID: scheduleDeletedID,
			DeletedAt:         deletedAt.Format(time.RFC3339),
			Scope:             scheduleScope(id),
		}),
		Metadata: metadata,
	}
}

func scheduleScope(id string) ScheduleScope {
	return ScheduleScope{ID: id}
}

func Channel(userRegisteredID string) string {
	return "schedule." + userRegisteredID
}
