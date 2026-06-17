package student

import (
	"time"

	"seek/internal/eventstore"
)

const (
	StudentCreated   = "StudentCreated"
	StudentRenamed   = "StudentRenamed"
	StudentCompleted = "StudentCompleted"
	StudentReopened  = "StudentReopened"
	StudentDeleted   = "StudentDeleted"
)

const (
	StudentIDField                    = "studentId"
	StudentUserRegisteredIDField      = "userRegisteredId"
	StudentTitleField                 = "title"
	StudentCreatedAtField             = "createdAt"
	StudentRenamedIDField             = "studentRenamedId"
	StudentRenamedAtField             = "renamedAt"
	StudentCompletedIDField           = "studentCompletedId"
	StudentCompletedAtField           = "completedAt"
	StudentReopenedIDField            = "studentReopenedId"
	StudentReopenedAtField            = "reopenedAt"
	StudentDeletedIDField             = "studentDeletedId"
	StudentDeletedAtField             = "deletedAt"
	StudentScopeIDField               = "scope.id"
)

type StudentCreatedEvent struct {
	ID				string				`json:"id"`
	FirstName	string				`json:"first_name"`
	CreatedAt string				`json:"createdAt"`
	Scope     StudentScope	`json:"scope"`
}

type StudentRenamedEvent struct {
	StudentRenamedID string    `json:"studentRenamedId"`
	Title         string    `json:"title"`
	RenamedAt     string    `json:"renamedAt"`
	Scope         StudentScope `json:"scope"`
}

type StudentCompletedEvent struct {
	StudentCompletedID string    `json:"studentCompletedId"`
	CompletedAt     string    `json:"completedAt"`
	Scope           StudentScope `json:"scope"`
}

type StudentReopenedEvent struct {
	StudentReopenedID string    `json:"studentReopenedId"`
	ReopenedAt     string    `json:"reopenedAt"`
	Scope          StudentScope `json:"scope"`
}

type StudentDeletedEvent struct {
	StudentDeletedID string    `json:"studentDeletedId"`
	DeletedAt     string    `json:"deletedAt"`
	Scope         StudentScope `json:"scope"`
}

type StudentScope struct {
	StudentID           string `json:"studentId"`
	UserRegisteredID string `json:"userRegisteredId"`
}

func NewStudentCreatedEvent(studentID, firstName string, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentID,
		EventType: StudentCreated,
		Data: eventstore.MustData(StudentCreatedEvent{
			ID:           studentID,
			FirstName:        firstName,
			CreatedAt:        createdAt.Format(time.RFC3339),
			Scope:            studentScope(studentID),
		}),
		Metadata: metadata,
	}
}

func NewStudentRenamedEvent(studentRenamedID, studentID, userRegisteredID, title string, renamedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentRenamedID,
		EventType: StudentRenamed,
		Data: eventstore.MustData(StudentRenamedEvent{
			StudentRenamedID: studentRenamedID,
			Title:         title,
			RenamedAt:     renamedAt.Format(time.RFC3339),
			Scope:         studentScope(studentID),
		}),
		Metadata: metadata,
	}
}

func NewStudentCompletedEvent(studentCompletedID, studentID, userRegisteredID string, completedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentCompletedID,
		EventType: StudentCompleted,
		Data: eventstore.MustData(StudentCompletedEvent{
			StudentCompletedID: studentCompletedID,
			CompletedAt:     completedAt.Format(time.RFC3339),
			Scope:           studentScope(studentID),
		}),
		Metadata: metadata,
	}
}

func NewStudentReopenedEvent(studentReopenedID, studentID string, reopenedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentReopenedID,
		EventType: StudentReopened,
		Data: eventstore.MustData(StudentReopenedEvent{
			StudentReopenedID: studentReopenedID,
			ReopenedAt:     reopenedAt.Format(time.RFC3339),
			Scope:          studentScope(studentID),
		}),
		Metadata: metadata,
	}
}

func NewStudentDeletedEvent(studentDeletedID, studentID string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentDeletedID,
		EventType: StudentDeleted,
		Data: eventstore.MustData(StudentDeletedEvent{
			StudentDeletedID: studentDeletedID,
			DeletedAt:     deletedAt.Format(time.RFC3339),
			Scope:         studentScope(studentID),
		}),
		Metadata: metadata,
	}
}

func studentScope(studentID string) StudentScope {
	return StudentScope{StudentID: studentID}
}

func Channel(userRegisteredID string) string {
	return "student." + userRegisteredID
}
