package events

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
	StudentIDField          = "student_id"
	StudentFirstNameField   = "first_name"
	StudentChosenNameField  = "chosen_name"
	StudentLastNameField    = "last_name"
	StudentGradeField       = "grade"
	StudentHomeroomField    = "homeroom"
	StudentCaseManagerField = "case_manager"
	StudentCreatedIDField   = "student_created_id"
	StudentCreatedAtField   = "created_at"
	StudentUpdatedIDField   = "student_updated_id"
	StudentUpdatedAtField   = "updated_at"
	StudentDeletedIDField   = "student_deleted_id"
	StudentDeletedAtField   = "deleted_at"
	StudentScopeIDField     = "scope.student_id"
)

type StudentCreatedEvent struct {
	StudentCreatedEventID string       `json:"student_created_id"`
	FirstName             string       `json:"first_name"`
	ChosenName            string       `json:"chosen_name"`
	LastName              string       `json:"last_name"`
	Grade                 int64        `json:"grade"`
	Homeroom              string       `json:"homeroom"`
	CaseManager           string       `json:"case_manager"`
	CreatedAt             string       `json:"created_at"`
	Scope                 StudentScope `json:"scope"`
}

type StudentUpdatedEvent struct {
	StudentUpdatedEventID string       `json:"student_updated_id"`
	FirstName             string       `json:"first_name"`
	ChosenName            string       `json:"chosen_name"`
	LastName              string       `json:"last_name"`
	Grade                 int64        `json:"grade"`
	Homeroom              string       `json:"homeroom"`
	CaseManager           string       `json:"case_manager"`
	UpdatedAt             string       `json:"updated_at"`
	Scope                 StudentScope `json:"scope"`
}

type StudentDeletedEvent struct {
	StudentDeletedEventID string       `json:"student_deleted_id"`
	DeletedAt             string       `json:"deleted_at"`
	Scope                 StudentScope `json:"scope"`
}

type StudentScope struct {
	ID string `json:"student_id"`
}

func NewStudentCreatedEvent(id, firstName, chosenName, lastName string, grade int64, homeroom, caseManager string, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: StudentCreated,
		Data: eventstore.MustData(StudentCreatedEvent{
			StudentCreatedEventID: id,
			FirstName:             firstName,
			ChosenName:            chosenName,
			LastName:              lastName,
			Grade:                 grade,
			Homeroom:              homeroom,
			CaseManager:           caseManager,
			CreatedAt:             createdAt.Format(time.RFC3339),
			Scope:                 studentScope(id),
		}),
		Metadata: metadata,
	}
}

func NewStudentUpdatedEvent(eventId, id, firstName, chosenName, lastName string, grade int64, homeroom, caseManager string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   eventId,
		EventType: StudentUpdated,
		Data: eventstore.MustData(StudentUpdatedEvent{
			StudentUpdatedEventID: id,
			FirstName:             firstName,
			ChosenName:            chosenName,
			LastName:              lastName,
			Grade:                 grade,
			Homeroom:              homeroom,
			CaseManager:           caseManager,
			UpdatedAt:             updatedAt.Format(time.RFC3339),
			Scope:                 studentScope(id),
		}),
		Metadata: metadata,
	}
}

func NewStudentDeletedEvent(studentDeletedID, id string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   studentDeletedID,
		EventType: StudentDeleted,
		Data: eventstore.MustData(StudentDeletedEvent{
			StudentDeletedEventID: studentDeletedID,
			DeletedAt:             deletedAt.Format(time.RFC3339),
			Scope:                 studentScope(id),
		}),
		Metadata: metadata,
	}
}

func studentScope(id string) StudentScope {
	return StudentScope{ID: id}
}

func Channel(id string) string {
	return "student." + id
}
