package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	EducatorAddedToPeriod     = "educator_added_to_period_event"
	EducatorRemovedFromPeriod = "educator_removed_from_period_event"
)

// EVENT FIELDS
const (
	EducatorAddedToPeriodIDField     = "educator_added_to_period_event_id"
	EducatorAddedToPeriodAtField     = "added_at"
	EducatorRemovedFromPeriodIDField = "educator_removed_from_period_event_id"
	EducatorRemovedFromPeriodAtField = "removed_at"
)

type EducatorAddedToPeriodEvent struct {
	EventID    string              `json:"educator_added_to_period_event_id"`
	PeriodID   string              `json:"period_id"`
	EducatorID string              `json:"educator_id"`
	AddedAt    time.Time           `json:"added_at"`
	Scope      PeriodEducatorScope `json:"scope"`
}

type EducatorRemovedFromPeriodEvent struct {
	EventID    string              `json:"educator_removed_from_period_event_id"`
	PeriodID   string              `json:"period_id"`
	EducatorID string              `json:"educator_id"`
	RemovedAt  time.Time           `json:"removed_at"`
	Scope      PeriodEducatorScope `json:"scope"`
}

type PeriodEducatorScope struct {
	PeriodID   string `json:"period_id"`
	EducatorID string `json:"educator_id"`
}

func NewEducatorAddedToPeriodEvent(
	eventID,
	periodID,
	educatorID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorAddedToPeriodEvent{
		EventID:    eventID,
		PeriodID:   periodID,
		EducatorID: educatorID,
		AddedAt:    addedAt,
		Scope:      periodEducatorScope(periodID, educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorAddedToPeriod,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewEducatorRemovedFromPeriodEvent(
	eventID,
	periodID,
	educatorID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := EducatorRemovedFromPeriodEvent{
		EventID:    eventID,
		PeriodID:   periodID,
		EducatorID: educatorID,
		RemovedAt:  removedAt,
		Scope:      periodEducatorScope(periodID, educatorID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorRemovedFromPeriod,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func periodEducatorScope(periodID, educatorID string) PeriodEducatorScope {
	return PeriodEducatorScope{
		PeriodID:   periodID,
		EducatorID: educatorID,
	}
}
