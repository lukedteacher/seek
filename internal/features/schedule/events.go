package schedule

import (
	"time"

	"seek/internal/eventstore"
)

const (
	ScheduleCreated       = "ScheduleCreated"
	ScheduleUpdated       = "ScheduleUpdated"
	PeriodAddedToSchedule = "PeriodAddedToSchedule"
	ScheduleDeleted       = "ScheduleDeleted"
)

const (
	ScheduleIDField               = "scheduleID"
	ScheduleUserRegisteredIDField = "userRegisteredId"
	ScheduleTitleField            = "title"
	ScheduleCreatedAtField        = "createdAt"
	ScheduleUpdatedIDField        = "scheduleUpdatedId"
	ScheduleUpdatedAtField        = "updatedAt"
	PeriodAddedToScheduleIDField  = "periodAddedToScheduleId"
	PeriodAddedToScheduleAtField  = "periodAddedToScheduleAt"
	ScheduleDeletedIDField        = "scheduleDeletedId"
	ScheduleDeletedAtField        = "deletedAt"
	ScheduleScopeIDField          = "scope.id"
)

type ScheduleCreatedEvent struct {
	ScheduleID string        `json:"scheduleID"`
	Title      string        `json:"title"`
	TeacherId  string        `json:"teacherID"`
	CreatedAt  string        `json:"createdAt"`
	Scope      ScheduleScope `json:"scope"`
}

type ScheduleUpdatedEvent struct {
	ScheduleUpdatedID string        `json:"scheduleUpdatedID"`
	Title             string        `json:"title"`
	TeacherId         string        `json:"teacherID"`
	UpdatedAt         string        `json:"updatedAt"`
	Scope             ScheduleScope `json:"scope"`
}

type PeriodAddedToScheduleEvent struct {
	PeriodAddedToScheduleID string        `json:"periodAddedToScheduleID"`
	ScheduleID              string        `json:"scheduleID"`
	PeriodID                string        `json:"periodID"`
	PeriodAddedToScheduleAt string        `json:"periodAddedToScheduleAt"`
	Scope                   ScheduleScope `json:"scope"`
}

type ScheduleDeletedEvent struct {
	ScheduleDeletedID string        `json:"scheduleDeletedID"`
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
			ScheduleID: id,
			Title:      title,
			TeacherId:  teacherId,
			CreatedAt:  createdAt.Format(time.RFC3339),
			Scope:      scheduleScope(id),
		}),
		Metadata: metadata,
	}
}

func NewScheduleUpdatedEvent(eventID, scheduleID, title, teacherId string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ScheduleUpdated,
		Data: eventstore.MustData(ScheduleUpdatedEvent{
			ScheduleUpdatedID: eventID,
			Title:             title,
			TeacherId:         teacherId,
			UpdatedAt:         updatedAt.Format(time.RFC3339),
			Scope:             scheduleScope(scheduleID),
		}),
		Metadata: metadata,
	}
}

func NewPeriodAddedToScheduleEvent(eventID, scheduleID string, periodID string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	event := PeriodAddedToScheduleEvent{
		PeriodAddedToScheduleID: eventID,
		ScheduleID:              scheduleID,
		PeriodID:                periodID,
		PeriodAddedToScheduleAt: updatedAt.Format(time.RFC3339),
		Scope:                   scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodAddedToSchedule,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
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
