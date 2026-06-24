package teacher

import (
	"time"

	"seek/internal/eventstore"
)

const (
	TeacherCreated = "TeacherCreated"
	TeacherUpdated = "TeacherUpdated"
	TeacherDeleted = "TeacherDeleted"
)

const (
	TeacherIDField               = "teacherId"
	TeacherUserRegisteredIDField = "userRegisteredId"
	TeacherFirstNameField        = "firstName"
	TeacherCreatedAtField        = "createdAt"
	TeacherUpdatedIDField        = "teacherUpdatedId"
	TeacherUpdatedAtField        = "updatedAt"
	TeacherDeletedIDField        = "teacherDeletedId"
	TeacherDeletedAtField        = "deletedAt"
	TeacherScopeIDField          = "scope.id"
)

type TeacherCreatedEvent struct {
	ID          string       `json:"id"`
	FirstName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	LastName    string       `json:"last_name"`
	CreatedAt   string       `json:"createdAt"`
	Scope       TeacherScope `json:"scope"`
}

type TeacherUpdatedEvent struct {
	TeacherUpdatedID string       `json:"teacherUpdatedId"`
	FirstName        string       `json:"first_name"`
	ChosenName       string       `json:"chosen_name"`
	LastName         string       `json:"last_name"`
	UpdatedAt        string       `json:"updatedAt"`
	Scope            TeacherScope `json:"scope"`
}

type TeacherDeletedEvent struct {
	TeacherDeletedID string       `json:"teacherDeletedId"`
	DeletedAt        string       `json:"deletedAt"`
	Scope            TeacherScope `json:"scope"`
}

type TeacherScope struct {
	ID               string `json:"id"`
	UserRegisteredID string `json:"userRegisteredId"`
}

func NewTeacherCreatedEvent(id, firstName, chosenName, lastName string, createdAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   id,
		EventType: TeacherCreated,
		Data: eventstore.MustData(TeacherCreatedEvent{
			ID:         id,
			FirstName:  firstName,
			ChosenName: chosenName,
			LastName:   lastName,
			CreatedAt:  createdAt.Format(time.RFC3339),
			Scope:      teacherScope(id),
		}),
		Metadata: metadata,
	}
}

func NewTeacherUpdatedEvent(teacherUpdatedID, id, firstName, chosenName, lastName string, updatedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   teacherUpdatedID,
		EventType: TeacherUpdated,
		Data: eventstore.MustData(TeacherUpdatedEvent{
			TeacherUpdatedID: teacherUpdatedID,
			FirstName:        firstName,
			ChosenName:       chosenName,
			LastName:         lastName,
			UpdatedAt:        updatedAt.Format(time.RFC3339),
			Scope:            teacherScope(id),
		}),
		Metadata: metadata,
	}
}

func NewTeacherDeletedEvent(teacherDeletedID, id string, deletedAt time.Time, metadata map[string]any) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   teacherDeletedID,
		EventType: TeacherDeleted,
		Data: eventstore.MustData(TeacherDeletedEvent{
			TeacherDeletedID: teacherDeletedID,
			DeletedAt:        deletedAt.Format(time.RFC3339),
			Scope:            teacherScope(id),
		}),
		Metadata: metadata,
	}
}

func teacherScope(id string) TeacherScope {
	return TeacherScope{ID: id}
}

func Channel(userRegisteredID string) string {
	return "teacher." + userRegisteredID
}
