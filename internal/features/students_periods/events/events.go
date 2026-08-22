package events

import (
	"time"

	"seek/internal/eventstore"
)

// event types
const (
	StudentAddedToPeriod     = "student_added_to_period_event"
	StudentRemovedFromPeriod = "student_removed_from_period_event"
)

// event fields
const (
	StudentAddedToPeriodIDField     = "student_added_to_period_event_id"
	StudentAddedToPeriodAtField     = "added_at"
	StudentRemovedFromPeriodIDField = "student_removed_from_period_event_id"
	StudentRemovedFromPeriodAtField = "removed_at"
)

type StudentAddedToPeriodEvent struct {
	EventID   string             `json:"student_added_to_period_event_id"`
	PeriodID  string             `json:"period_id"`
	StudentID string             `json:"student_id"`
	AddedAt   time.Time          `json:"added_at"`
	Scope     PeriodStudentScope `json:"scope"`
}

type StudentRemovedFromPeriodEvent struct {
	EventID   string             `json:"student_removed_from_period_event_id"`
	PeriodID  string             `json:"period_id"`
	StudentID string             `json:"student_id"`
	RemovedAt time.Time          `json:"removed_at"`
	Scope     PeriodStudentScope `json:"scope"`
}

type PeriodStudentScope struct {
	PeriodID  string `json:"period_id"`
	StudentID string `json:"student_id"`
}

func NewStudentAddedToPeriodEvent(
	eventID,
	periodID,
	studentID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentAddedToPeriodEvent{
		EventID:   eventID,
		PeriodID:  periodID,
		StudentID: studentID,
		AddedAt:   addedAt,
		Scope:     periodStudentScope(periodID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentAddedToPeriod,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentRemovedFromPeriodEvent(
	eventID,
	periodID,
	studentID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentRemovedFromPeriodEvent{
		EventID:   eventID,
		PeriodID:  periodID,
		StudentID: studentID,
		RemovedAt: removedAt,
		Scope:     periodStudentScope(periodID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentRemovedFromPeriod,
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
