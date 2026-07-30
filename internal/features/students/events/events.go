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

// student event id & scope fields
const (
	FieldStudentCreatedID  = "student_created_event_id"
	FieldStudentUpdatedID  = "student_updated_event_id"
	FieldStudentArchivedID = "student_archived_event_id"
	FieldStudentDeletedID  = "student_deleted_event_id"
	FieldStudentScopeID    = "scope.student_id"
)

// student event fields
const (
	FieldStudentID          = "student_id"
	FieldStudentMARSSID     = "marss_id"
	FieldStudentGivenName   = "given_name"
	FieldStudentChosenName  = "chosen_name"
	FieldStudentFamilyName  = "family_name"
	FieldStudentEmail       = "email"
	FieldStudentUsername    = "username"
	FieldStudentGrade       = "grade"
	FieldStudentHomeroom    = "homeroom"
	FieldStudentCaseManager = "case_manager"
	FieldStudentCreatedAt   = "created_at"
	FieldStudentUpdatedAt   = "updated_at"
	FieldStudentArchivedAt  = "archived_at"
	FieldStudentDeletedAt   = "deleted_at"
)

type StudentCreatedEvent struct {
	EventID     string       `json:"student_created_event_id"`
	MARSSID     string       `json:"marss_id"`
	GivenName   string       `json:"given_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"family_name"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	Grade       int          `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	CreatedAt   string       `json:"created_at"`
	Scope       StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	EventID     string       `json:"student_updated_event_id"`
	MARSSID     string       `json:"marss_id"`
	GivenName   string       `json:"given_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"family_name"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
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
	marssID,
	givenName,
	chosenName,
	familyName,
	email,
	username string,
	grade int,
	homeroom,
	caseManager string,
	createdAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentCreatedEvent{
		EventID:     eventID,
		MARSSID:     marssID,
		GivenName:   givenName,
		ChosenName:  chosenName,
		FamilyName:  familyName,
		Email:       email,
		Username:    username,
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
	marssID,
	givenName,
	chosenName,
	familyName,
	email,
	username string,
	grade int,
	homeroom,
	caseManager string,
	updatedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := StudentUpdatedEvent{
		EventID:     eventID,
		MARSSID:     marssID,
		GivenName:   givenName,
		ChosenName:  chosenName,
		FamilyName:  familyName,
		Email:       email,
		Username:    username,
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
