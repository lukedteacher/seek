package student

import (
	"time"

	"seek/internal/eventstore"
)

const (
	StudentCreated = "StudentCreated"
	StudentUpdated = "StudentUpdated"
	StudentDeleted = "StudentDeleted"
)

const (
	StudentIDField               = "studentId"
	StudentUserRegisteredIDField = "userRegisteredId"
	StudentFirstNameField        = "firstName"
	StudentCreatedAtField        = "createdAt"
	StudentUpdatedIDField        = "studentUpdatedId"
	StudentUpdatedAtField        = "updatedAt"
	StudentDeletedIDField        = "studentDeletedId"
	StudentDeletedAtField        = "deletedAt"
	StudentScopeIDField          = "scope.id"
)

type StudentCreatedEvent struct {
	ID          string       `json:"id"`
	FirstName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	LastName    string       `json:"last_name"`
	Grade       int64        `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	CreatedAt   string       `json:"createdAt"`
	Scope       StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	StudentUpdatedID string       `json:"studentUpdatedId"`
	FirstName        string       `json:"first_name"`
	ChosenName       string       `json:"chosen_name"`
	LastName         string       `json:"last_name"`
	UpdatedAt        string       `json:"updatedAt"`
	Scope            StudentScope `json:"scope"`
}

type StudentDeletedEvent struct {
	StudentDeletedID string       `json:"studentDeletedId"`
	DeletedAt        string       `json:"deletedAt"`
	Scope            StudentScope `json:"scope"`
}

type StudentScope struct {
	ID               string `json:"id"`
	UserRegisteredID string `json:"userRegisteredId"`
}

func NewStudentCreatedEvent(id, firstName, chosenName, lastName string, grade int64, homeroom, caseManager string, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: StudentCreated,
		Data: eventstore.MustData(StudentCreatedEvent{
			ID:          id,
			FirstName:   firstName,
			ChosenName:  chosenName,
			LastName:    lastName,
			Grade:       grade,
			Homeroom:    homeroom,
			CaseManager: caseManager,
			CreatedAt:   createdAt.Format(time.RFC3339),
			Scope:       studentScope(id),
		}),
		Metadata: metadata,
	}
}

func NewStudentUpdatedEvent(eventId, id, firstName string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   eventId,
		EventType: StudentUpdated,
		Data: eventstore.MustData(StudentUpdatedEvent{
			StudentUpdatedID: id,
			FirstName:        firstName,
			UpdatedAt:        updatedAt.Format(time.RFC3339),
			Scope:            studentScope(id),
		}),
		Metadata: metadata,
	}
}

func NewStudentDeletedEvent(studentDeletedID, id string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentDeletedID,
		EventType: StudentDeleted,
		Data: eventstore.MustData(StudentDeletedEvent{
			StudentDeletedID: studentDeletedID,
			DeletedAt:        deletedAt.Format(time.RFC3339),
			Scope:            studentScope(id),
		}),
		Metadata: metadata,
	}
}

func studentScope(id string) StudentScope {
	return StudentScope{ID: id}
}

func Channel(userRegisteredID string) string {
	return "student." + userRegisteredID
}
