package events

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
)

type eventType = eventstore.EventType

// event types
const (
	EventPeriodCreated  eventType = "period_created_event"
	EventPeriodUpdated  eventType = "period_updated_event"
	EventPeriodArchived eventType = "period_archived_event"
	EventPeriodDeleted  eventType = "period_deleted_event"
)

// event ID fields
const (
	FieldPeriodCreatedEventID  = "period_created_event_id"
	FieldPeriodUpdatedEventID  = "period_updated_event_id"
	FieldPeriodArchivedEventID = "period_archived_event_id"
	FieldPeriodDeletedEventID  = "period_deleted_event_id"
)

// event data fields
const (
	FieldPeriodID          = "period_id"
	FieldPeriodTitle       = "title"
	FieldPeriodServiceType = "service_type"
	FieldPeriodStartTime   = "start_time"
	FieldPeriodDuration    = "duration"
	FieldPeriodDaysBitmask = "days_bitmask"
	FieldPeriodCreatedAt   = "created_at"
	FieldPeriodUpdatedAt   = "updated_at"
	FieldPeriodArchivedAt  = "archived_at"
	FieldPeriodDeletedAt   = "deleted_at"
)

// event scope fields
const (
	FieldPeriodScopeID = "scope.period_id"
)

type PeriodCreatedEvent struct {
	EventID     string                   `json:"period_created_event_id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	CreatedAt   string                   `json:"created_at"`
	Scope       PeriodScope              `json:"scope"`
}

type PeriodUpdatedEvent struct {
	EventID     string                   `json:"period_updated_event_id"`
	Title       string                   `json:"title"`
	ServiceType sharedmodels.ServiceType `json:"service_type"`
	StartTime   sharedmodels.TimeOnly    `json:"start_time"`
	Duration    int                      `json:"duration"`
	DaysBitmask sharedmodels.DaysBitmask `json:"days_bitmask"`
	UpdatedAt   string                   `json:"updated_at"`
	Scope       PeriodScope              `json:"scope"`
}

type PeriodArchivedEvent struct {
	EventID    string      `json:"period_archived_event_id"`
	ArchivedAt string      `json:"archived_at"`
	Scope      PeriodScope `json:"scope"`
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
	eventID,
	title string,
	serviceType sharedmodels.ServiceType,
	startTime sharedmodels.TimeOnly,
	duration int,
	daysBitmask sharedmodels.DaysBitmask,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodCreatedEvent{
		EventID:     eventID,
		Title:       title,
		ServiceType: serviceType,
		StartTime:   startTime,
		Duration:    duration,
		DaysBitmask: daysBitmask,
		CreatedAt:   createdAt.Format(time.RFC3339),
		Scope:       periodScope(eventID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventPeriodCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodUpdatedEvent(
	eventID,
	periodID,
	title string,
	serviceType sharedmodels.ServiceType,
	startTime sharedmodels.TimeOnly,
	duration int,
	daysBitmask sharedmodels.DaysBitmask,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodUpdatedEvent{
		EventID:     eventID,
		Title:       title,
		ServiceType: serviceType,
		StartTime:   startTime,
		Duration:    duration,
		DaysBitmask: daysBitmask,
		UpdatedAt:   updatedAt.Format(time.RFC3339),
		Scope:       periodScope(periodID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventPeriodUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodArchivedEvent(
	eventID,
	periodID string,
	archivedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodArchivedEvent{
		EventID:    eventID,
		ArchivedAt: archivedAt.Format(time.RFC3339),
		Scope:      periodScope(periodID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventPeriodArchived,
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
		EventType: EventPeriodDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func periodScope(id string) PeriodScope {
	return PeriodScope{ID: id}
}

func Channel(id string) string {
	return "periods." + id
}

func ChannelAll() string {
	return "periods.>"
}
