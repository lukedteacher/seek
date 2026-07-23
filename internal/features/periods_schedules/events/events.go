package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	PeriodScheduleAdded   = "PeriodScheduleAdded"
	PeriodScheduleRemoved = "PeriodScheduleRemoved"
)

// EVENT FIELDS
const (
	PeriodScheduleAddedIDField   = "period_schedule_added_event_id"
	PeriodScheduleAddedAtField   = "added_at"
	PeriodScheduleRemovedIDField = "period_schedule_removed_event_id"
	PeriodScheduleRemovedAtField = "removed_at"
)

type PeriodScheduleAddedEvent struct {
	EventID string              `json:"period_schedule_added_event_id"`
	AddedAt string              `json:"added_at"`
	Scope   PeriodScheduleScope `json:"scope"`
}

type PeriodScheduleRemovedEvent struct {
	EventID   string              `json:"period_schedule_removed_event_id"`
	RemovedAt string              `json:"removed_at"`
	Scope     PeriodScheduleScope `json:"scope"`
}

type PeriodScheduleScope struct {
	PeriodID   string `json:"period_id"`
	ScheduleID string `json:"schedule_id"`
}

func NewPeriodScheduleAddedEvent(
	eventID,
	periodID,
	scheduleID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodScheduleAddedEvent{
		AddedAt: addedAt.Format(time.RFC3339),
		Scope:   periodScheduleScope(periodID, scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodScheduleAdded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodScheduleRemovedEvent(
	eventID,
	periodID,
	scheduleID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodScheduleRemovedEvent{
		EventID:   eventID,
		RemovedAt: removedAt.Format(time.RFC3339),
		Scope:     periodScheduleScope(periodID, scheduleID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodScheduleRemoved,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func periodScheduleScope(periodID, scheduleID string) PeriodScheduleScope {
	return PeriodScheduleScope{
		PeriodID:   periodID,
		ScheduleID: scheduleID,
	}
}
