package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	PeriodCreated = "PeriodCreated"
	PeriodUpdated = "PeriodUpdated"
	PeriodDeleted = "PeriodDeleted"
)

// EVENT FIELDS
const (
	PeriodIDField        = "period_id"
	PeriodTitleField     = "title"
	PeriodStartTimeField = "start_time"
	PeriodDurationField  = "duration"
	PeriodDaysField      = "days"
	PeriodCreatedIDField = "period_created_event_id"
	PeriodCreatedAtField = "created_at"
	PeriodUpdatedIDField = "period_updated_event_id"
	PeriodUpdatedAtField = "updated_at"
	PeriodDeletedIDField = "period_deleted_event_id"
	PeriodDeletedAtField = "deleted_at"
	PeriodScopeIDField   = "scope.period_id"
)

// includes period scope which may be redundant
// since event ID is the same as the period for created
type PeriodCreatedEvent struct {
	EventID   string      `json:"period_created_event_id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int64       `json:"duration"`
	Days      int64       `json:"days"`
	CreatedAt string      `json:"created_at"`
	Scope     PeriodScope `json:"scope"`
}

type PeriodUpdatedEvent struct {
	EventID   string      `json:"period_updated_event_id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int64       `json:"duration"`
	Days      int64       `json:"days"`
	UpdatedAt string      `json:"updated_at"`
	Scope     PeriodScope `json:"scope"`
}

type PeriodDeletedEvent struct {
	EventID   string      `json:"period_deleted_event_id"`
	DeletedAt string      `json:"deleted_at"`
	Scope     PeriodScope `json:"scope"`
}

type PeriodScope struct {
	ID string `json:"period_id"`
}

func NewPeriodCreatedEvent(
	periodID,
	title,
	startTime string,
	duration,
	days int64,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodCreatedEvent{
		EventID:   periodID,
		Title:     title,
		StartTime: startTime,
		Duration:  duration,
		Days:      days,
		CreatedAt: createdAt.Format(time.RFC3339),
		Scope:     periodScope(periodID),
	}
	return eventstore.DomainEvent{
		EventID:   periodID,
		EventType: PeriodCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodUpdatedEvent(
	eventID,
	periodID,
	title,
	startTime string,
	duration,
	days int64,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodUpdatedEvent{
		EventID:   eventID,
		Title:     title,
		StartTime: startTime,
		Duration:  duration,
		Days:      days,
		UpdatedAt: updatedAt.Format(time.RFC3339),
		Scope:     periodScope(periodID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodDeletedEvent(
	eventID,
	periodID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     periodScope(periodID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func periodScope(id string) PeriodScope {
	return PeriodScope{ID: id}
}

func Channel(id string) string {
	return "period." + id
}
