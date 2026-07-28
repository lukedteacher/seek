package events

import (
	"time"

	"seek/internal/eventstore"
)

// student event types
const (
	StudentCreated  = "StudentCreated"
	StudentUpdated  = "StudentUpdated"
	StudentArchived = "StudentArchived"
	StudentDeleted  = "StudentDeleted"
)

const (
	StudentCreatedIDField  = "student_created_event_id"
	StudentUpdatedIDField  = "student_updated_event_id"
	StudentArchivedIDField = "student_archived_event_id"
	StudentDeletedIDField  = "student_deleted_event_id"
)

// student event fields
const (
	StudentIDField          = "student_id"
	StudentGivenNameField   = "given_name"
	StudentChosenNameField  = "chosen_name"
	StudentFamilyNameField  = "family_name"
	StudentGradeField       = "grade"
	StudentHomeroomField    = "homeroom"
	StudentCaseManagerField = "case_manager"
	StudentCreatedAtField   = "created_at"
	StudentUpdatedAtField   = "updated_at"
	StudentArchivedAtField  = "archived_at"
	StudentDeletedAtField   = "deleted_at"
	StudentScopeIDField     = "scope.student_id"
)

type StudentCreatedEvent struct {
	EventID     string       `json:"student_created_event_id"`
	GivenName   string       `json:"given_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"family_name"`
	Grade       int          `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	CreatedAt   string       `json:"created_at"`
	Scope       StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	EventID     string       `json:"student_updated_event_id"`
	GivenName   string       `json:"given_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"family_name"`
	Grade       int          `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	UpdatedAt   string       `json:"updated_at"`
	Scope       StudentScope `json:"scope"`
}

type StudentArchivedEvent struct {
	EventID    string       `json:"student_archived_event_id"`
	ArchivedAt string       `json:"archived_at"`
	Scope      StudentScope `json:"scope"`
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
	eventID,
	givenName,
	chosenName,
	familyName string,
	grade int,
	homeroom,
	caseManager string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentCreatedEvent{
		EventID:     eventID,
		GivenName:   givenName,
		ChosenName:  chosenName,
		FamilyName:  familyName,
		Grade:       grade,
		Homeroom:    homeroom,
		CaseManager: caseManager,
		CreatedAt:   createdAt.Format(time.RFC3339),
		Scope:       studentScope(eventID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentCreated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewStudentUpdatedEvent(
	eventID,
	studentID,
	givenName,
	chosenName,
	familyName string,
	grade int,
	homeroom,
	caseManager string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentUpdatedEvent{
		EventID:     eventID,
		GivenName:   givenName,
		ChosenName:  chosenName,
		FamilyName:  familyName,
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

func NewStudentArchivedEvent(
	eventID,
	studentID string,
	archivedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentArchivedEvent{
		EventID:    eventID,
		ArchivedAt: archivedAt.Format(time.RFC3339),
		Scope:      studentScope(studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: StudentArchived,
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
