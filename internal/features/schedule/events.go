package schedule

import (
	"time"

	"seek/internal/eventstore"
)

const (
	ScheduleCreated       = "ScheduleCreated"
	ScheduleUpdated       = "ScheduleUpdated"
	SchedulePeriodAdded   = "SchedulePeriodAdded"
	SchedulePeriodRemoved = "SchedulePeriodRemoved"
	ScheduleDeleted       = "ScheduleDeleted"
)

const (
	ScheduleIDField               = "scheduleID"
	ScheduleUserRegisteredIDField = "userRegisteredId"
	ScheduleTitleField            = "title"
	ScheduleCreatedAtField        = "createdAt"
	ScheduleUpdatedIDField        = "scheduleUpdatedId"
	ScheduleUpdatedAtField        = "updatedAt"
	SchedulePeriodAddedIDField    = "schedulePeriodAddedID"
	SchedulePeriodAddedAtField    = "periodAddedAt"
	SchedulePeriodRemovedIDField  = "schedulePeriodRemovedID"
	SchedulePeriodRemovedAtField  = "periodRemovedAt"
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

type SchedulePeriodAddedEvent struct {
	SchedulePeriodAddedID string        `json:"schedulePeriodAddedID"`
	ScheduleID            string        `json:"scheduleID"`
	PeriodID              string        `json:"periodID"`
	PeriodAddedAt         string        `json:"periodAddedAt"`
	Scope                 ScheduleScope `json:"scope"`
}

type SchedulePeriodRemovedEvent struct {
	SchedulePeriodRemovedID string        `json:"schedulePeriodRemovedID"`
	ScheduleID              string        `json:"scheduleID"`
	PeriodID                string        `json:"periodID"`
	PeriodRemovedAt         string        `json:"periodRemovedAt"`
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

func NewSchedulePeriodAddedEvent(eventID, scheduleID string, periodID string, addedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	event := SchedulePeriodAddedEvent{
		SchedulePeriodAddedID: eventID,
		ScheduleID:            scheduleID,
		PeriodID:              periodID,
		PeriodAddedAt:         addedAt.Format(time.RFC3339),
		Scope:                 scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: SchedulePeriodAdded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewSchedulePeriodRemovedEvent(eventID, scheduleID string, periodID string, removedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	event := SchedulePeriodRemovedEvent{
		SchedulePeriodRemovedID: eventID,
		ScheduleID:              scheduleID,
		PeriodID:                periodID,
		PeriodRemovedAt:         removedAt.Format(time.RFC3339),
		Scope:                   scheduleScope(scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: SchedulePeriodRemoved,
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
