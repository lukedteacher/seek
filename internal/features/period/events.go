package period

import (
	"time"

	"seek/internal/eventstore"
)

const (
	PeriodCreated   = "PeriodCreated"
	PeriodRenamed   = "PeriodRenamed"
	PeriodDeleted   = "PeriodDeleted"
)

const (
	PeriodIDField                    = "periodId"
	PeriodUserRegisteredIDField      = "userRegisteredId"
	PeriodTitleField                 = "title"
	PeriodCreatedAtField             = "createdAt"
	PeriodRenamedIDField             = "periodRenamedId"
	PeriodRenamedAtField             = "renamedAt"
	PeriodDeletedIDField             = "periodDeletedId"
	PeriodDeletedAtField             = "deletedAt"
	PeriodScopeIDField               = "scope.id"
)

type PeriodCreatedEvent struct {
	ID        string			`json:"id"`
	Title     string			`json:"title"`
	StartTime string		  `json:"start_time"`
	Duration  int64				`json:"duration"`
	Days      int64       `json:"days"`
	CreatedAt string			`json:"createdAt"`
	Scope     PeriodScope	`json:"scope"`
}

type PeriodRenamedEvent struct {
	PeriodRenamedID string    `json:"periodRenamedId"`
	Title         string    `json:"title"`
	RenamedAt     string    `json:"renamedAt"`
	Scope         PeriodScope `json:"scope"`
}

type PeriodDeletedEvent struct {
	PeriodDeletedID string    `json:"periodDeletedId"`
	DeletedAt     string    `json:"deletedAt"`
	Scope         PeriodScope `json:"scope"`
}

type PeriodScope struct {
	ID           string `json:"id"`
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

func NewPeriodRenamedEvent(periodRenamedID, id, title string, renamedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   periodRenamedID,
		EventType: PeriodRenamed,
		Data: eventstore.MustData(PeriodRenamedEvent{
			PeriodRenamedID: periodRenamedID,
			Title:         title,
			RenamedAt:     renamedAt.Format(time.RFC3339),
			Scope:         periodScope(id),
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
			DeletedAt:     deletedAt.Format(time.RFC3339),
			Scope:         periodScope(id),
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
