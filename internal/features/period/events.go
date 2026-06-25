package period

import (
	"time"

	"seek/internal/eventstore"
)

const (
	PeriodCreated = "PeriodCreated"
	PeriodUpdated = "PeriodUpdated"
	PeriodDeleted = "PeriodDeleted"
)

const (
	PeriodIDField               = "periodId"
	PeriodUserRegisteredIDField = "userRegisteredId"
	PeriodTitleField            = "title"
	PeriodCreatedAtField        = "createdAt"
	PeriodUpdatedIDField        = "periodUpdatedId"
	PeriodUpdatedAtField        = "updatedAt"
	PeriodDeletedIDField        = "periodDeletedId"
	PeriodDeletedAtField        = "deletedAt"
	PeriodScopeIDField          = "scope.id"
)

type PeriodCreatedEvent struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int64       `json:"duration"`
	Days      int64       `json:"days"`
	CreatedAt string      `json:"createdAt"`
	Scope     PeriodScope `json:"scope"`
}

type PeriodUpdatedEvent struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	StartTime string      `json:"start_time"`
	Duration  int64       `json:"duration"`
	Days      int64       `json:"days"`
	UpdatedAt string      `json:"updatedAt"`
	Scope     PeriodScope `json:"scope"`
}

type PeriodDeletedEvent struct {
	PeriodDeletedID string      `json:"periodDeletedId"`
	DeletedAt       string      `json:"deletedAt"`
	Scope           PeriodScope `json:"scope"`
}

type PeriodScope struct {
	ID               string `json:"id"`
	UserRegisteredID string `json:"userRegisteredId"`
}

func NewPeriodCreatedEvent(id, title, startTime string, duration, days int64, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: PeriodCreated,
		Data: eventstore.MustData(PeriodCreatedEvent{
			ID:        id,
			Title:     title,
			StartTime: startTime,
			Duration:  duration,
			Days:      days,
			CreatedAt: createdAt.Format(time.RFC3339),
			Scope:     periodScope(id),
		}),
		Metadata: metadata,
	}
}

func NewPeriodUpdatedEvent(eventID, periodID, title, startTime string, duration, days int64, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodUpdated,
		Data: eventstore.MustData(PeriodUpdatedEvent{
			ID:        eventID,
			Title:     title,
			StartTime: startTime,
			Duration:  duration,
			Days:      days,
			UpdatedAt: updatedAt.Format(time.RFC3339),
			Scope:     periodScope(periodID),
		}),
		Metadata: metadata,
	}
}

func NewPeriodDeletedEvent(periodDeletedID, id string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   periodDeletedID,
		EventType: PeriodDeleted,
		Data: eventstore.MustData(PeriodDeletedEvent{
			PeriodDeletedID: periodDeletedID,
			DeletedAt:       deletedAt.Format(time.RFC3339),
			Scope:           periodScope(id),
		}),
		Metadata: metadata,
	}
}

func periodScope(id string) PeriodScope {
	return PeriodScope{ID: id}
}

func Channel(userRegisteredID string) string {
	return "period." + userRegisteredID
}
