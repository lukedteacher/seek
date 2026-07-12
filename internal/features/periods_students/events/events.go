package events

import (
	"time"

	"seek/internal/eventstore"
)

const (
	PeriodCreated        = "PeriodCreated"
	PeriodDeleted        = "PeriodDeleted"
	StudentCreated       = "StudentCreated"
	StudentDeleted       = "StudentDeleted"
	PeriodStudentAdded   = "PeriodStudentAdded"
	PeriodStudentRemoved = "PeriodStudentRemoved"
)

const (
	PeriodStudentAddedIDField   = "period_student_added_id"
	PeriodStudentAddedAtField   = "added_at"
	PeriodStudentRemovedIDField = "period_student_removed_id"
	PeriodStudentRemovedAtField = "removed_at"
)

type PeriodStudentAddedEvent struct {
	PeriodStudentAddedID string             `json:"period_student_added_id"`
	PeriodID             string             `json:"period_id"`
	StudentID            string             `json:"student_id"`
	AddedAt              string             `json:"added_at"`
	Scope                PeriodStudentScope `json:"scope"`
}

type PeriodStudentRemovedEvent struct {
	PeriodStudentRemovedID string             `json:"period_student_removed_id"`
	PeriodID               string             `json:"period_id"`
	StudentID              string             `json:"student_id"`
	RemovedAt              string             `json:"removed_at"`
	Scope                  PeriodStudentScope `json:"scope"`
}

type PeriodStudentScope struct {
	PeriodID  string `json:"period_id"`
	StudentID string `json:"student_id"`
}

func NewPeriodStudentAddedEvent(eventID, periodID, studentID string, addedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	event := PeriodStudentAddedEvent{
		PeriodStudentAddedID: eventID,
		PeriodID:             periodID,
		StudentID:            studentID,
		AddedAt:              addedAt.Format(time.RFC3339),
		Scope:                periodStudentScope(periodID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: PeriodStudentAdded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewPeriodStudentRemovedEvent(eventID, periodID, studentID string, removedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	event := PeriodStudentRemovedEvent{
		PeriodStudentRemovedID: eventID,
		PeriodID:               periodID,
		StudentID:              studentID,
		RemovedAt:              removedAt.Format(time.RFC3339),
		Scope:                  periodStudentScope(periodID, studentID),
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

func Channel(id string) string {
	return "periods_students." + id
}
