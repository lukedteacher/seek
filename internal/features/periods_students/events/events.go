package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	PeriodStudentAdded   = "PeriodStudentAdded"
	PeriodStudentRemoved = "PeriodStudentRemoved"
)

// EVENT FIELDS
const (
	PeriodStudentAddedIDField   = "period_student_added_event_id"
	PeriodStudentAddedAtField   = "added_at"
	PeriodStudentRemovedIDField = "period_student_removed_event_id"
	PeriodStudentRemovedAtField = "removed_at"
)

type PeriodStudentAddedEvent struct {
	EventID string             `json:"period_student_added_event_id"`
	AddedAt string             `json:"added_at"`
	Scope   PeriodStudentScope `json:"scope"`
}

type PeriodStudentRemovedEvent struct {
	EventID   string             `json:"period_student_removed_event_id"`
	RemovedAt string             `json:"removed_at"`
	Scope     PeriodStudentScope `json:"scope"`
}

type PeriodStudentScope struct {
	PeriodID  string `json:"period_id"`
	StudentID string `json:"student_id"`
}

func NewPeriodStudentAddedEvent(
	eventID,
	periodID,
	studentID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodStudentAddedEvent{
		EventID: eventID,
		AddedAt: addedAt.Format(time.RFC3339),
		Scope:   periodStudentScope(periodID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodStudentAdded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodStudentRemovedEvent(
	eventID,
	periodID,
	studentID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := PeriodStudentRemovedEvent{
		EventID:   eventID,
		RemovedAt: removedAt.Format(time.RFC3339),
		Scope:     periodStudentScope(periodID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodStudentRemoved,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func periodStudentScope(periodID, studentID string) PeriodStudentScope {
	return PeriodStudentScope{
		PeriodID:  periodID,
		StudentID: studentID,
	}
}
