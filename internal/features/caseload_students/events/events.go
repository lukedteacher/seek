package events

import (
	"time"

	"seek/internal/eventstore"
)

// case manager student event types

const (
	StudentAddedToCaseload     = "student_added_to_caseload_event"
	StudentRemovedFromCaseload = "student_removed_from_caseload_event"
)

// case manager student event fields
const (
	FieldStudentAddedToCaseloadEventID     = "student_added_to_caseload_event_id"
	FieldStudentAddedToCaseloadAt          = "added_at"
	FieldStudentRemovedFromCaseloadEventID = "student_removed_from_caseload_event_id"
	FieldStudentRemovedFromCaseloadAt      = "removed_at"
)

type StudentAddedToCaseloadEvent struct {
	EventID    string                  `json:"student_added_to_caseload_event_id"`
	EducatorID string                  `json:"educator_id"`
	StudentID  string                  `json:"student_id"`
	AddedAt    time.Time               `json:"added_at"`
	Scope      CaseManagerStudentScope `json:"scope"`
}

type StudentRemovedFromCaseloadEvent struct {
	EventID    string                  `json:"student_removed_from_caseload_event_id"`
	EducatorID string                  `json:"educator_id"`
	StudentID  string                  `json:"student_id"`
	RemovedAt  time.Time               `json:"removed_at"`
	Scope      CaseManagerStudentScope `json:"scope"`
}

type CaseManagerStudentScope struct {
	EducatorID string `json:"educator_id"`
	StudentID  string `json:"student_id"`
}

func NewStudentAddedToCaseloadEvent(
	eventID,
	educatorID,
	studentID string,
	addedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentAddedToCaseloadEvent{
		EventID:    eventID,
		EducatorID: educatorID,
		StudentID:  studentID,
		AddedAt:    addedAt,
		Scope:      caseManagerStudentScope(educatorID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentAddedToCaseload,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentRemovedFromCaseloadEvent(
	eventID,
	educatorID,
	studentID string,
	removedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentRemovedFromCaseloadEvent{
		EventID:    eventID,
		EducatorID: educatorID,
		StudentID:  studentID,
		RemovedAt:  removedAt,
		Scope:      caseManagerStudentScope(educatorID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentRemovedFromCaseload,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func caseManagerStudentScope(educatorID, studentID string) CaseManagerStudentScope {
	return CaseManagerStudentScope{
		EducatorID: educatorID,
		StudentID:  studentID,
	}
}
