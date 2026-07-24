package events

import (
	"time"

	"seek/internal/eventstore"
)

// EVENT NAMES
const (
	StudentCreated  = "StudentCreated"
	StudentUpdated  = "StudentUpdated"
	StudentArchived = "StudentArchived"
	StudentDeleted  = "StudentDeleted"
)

// EVENT FIELDS
const (
	StudentIDField          = "student_id"
	StudentGivenNameField   = "first_name"
	StudentChosenNameField  = "chosen_name"
	StudentFamilyNameField  = "last_name"
	StudentGradeField       = "grade"
	StudentHomeroomField    = "homeroom"
	StudentCaseManagerField = "case_manager"
	StudentCreatedIDField   = "student_created_event_id"
	StudentCreatedAtField   = "created_at"
	StudentUpdatedIDField   = "student_updated_event_id"
	StudentUpdatedAtField   = "updated_at"
	StudentDeletedIDField   = "student_deleted_event_id"
	StudentDeletedAtField   = "deleted_at"
	StudentScopeIDField     = "scope.student_id"
)

type StudentCreatedEvent struct {
	EventID     string       `json:"student_created_event_id"`
	GivenName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"last_name"`
	Grade       int64        `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	CreatedAt   string       `json:"created_at"`
	Scope       StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	EventID     string       `json:"student_updated_event_id"`
	GivenName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"last_name"`
	Grade       int64        `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	UpdatedAt   string       `json:"updated_at"`
	Scope       StudentScope `json:"scope"`
}

type StudentDeletedEvent struct {
	EventID   string       `json:"student_deleted_event_id"`
	DeletedAt string       `json:"deleted_at"`
	Scope     StudentScope `json:"scope"`
}

type StudentScope struct {
	ID string `json:"student_id"`
}

func NewStudentCreatedEvent(
	studentID,
	firstName,
	chosenName,
	lastName string,
	grade int64,
	homeroom,
	caseManager string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentCreatedEvent{
		EventID:     studentID,
		GivenName:   firstName,
		ChosenName:  chosenName,
		FamilyName:  lastName,
		Grade:       grade,
		Homeroom:    homeroom,
		CaseManager: caseManager,
		CreatedAt:   createdAt.Format(time.RFC3339),
		Scope:       studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   studentID,
		EventType: StudentCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentUpdatedEvent(
	eventID,
	studentID,
	firstName,
	chosenName,
	lastName string,
	grade int64,
	homeroom,
	caseManager string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentUpdatedEvent{
		EventID:     eventID,
		GivenName:   firstName,
		ChosenName:  chosenName,
		FamilyName:  lastName,
		Grade:       grade,
		Homeroom:    homeroom,
		CaseManager: caseManager,
		UpdatedAt:   updatedAt.Format(time.RFC3339),
		Scope:       studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentDeletedEvent(
	eventID,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentDeletedEvent{
		EventID:   eventID,
		DeletedAt: deletedAt.Format(time.RFC3339),
		Scope:     studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func studentScope(id string) StudentScope {
	return StudentScope{ID: id}
}

func Channel(id string) string {
	return "students." + id
}

func ChannelAll() string {
	return "students.>"
}
