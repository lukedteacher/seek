package teacher

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type UpdateTeacherCommand struct {
	UserRegisteredID string
	TeacherID        string
	FirstName        string
	ChosenName       string
	LastName         string
	Metadata         CommandMetadata
}

type UpdateTeacherResult struct {
	TeacherUpdatedID string
	Skipped          bool
}

func UpdateTeacherCommandHandler(ctx context.Context, command UpdateTeacherCommand, saver eventstore.Saver, retriever eventstore.Retriever) (UpdateTeacherResult, error) {
	model, err := loadUpdateTeacherContext(ctx, retriever, command.TeacherID)
	if err != nil {
		return UpdateTeacherResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return UpdateTeacherResult{}, err
	}
	if model.firstName == command.FirstName && model.chosenName == command.ChosenName && model.lastName == command.LastName {
		return UpdateTeacherResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewTeacherUpdatedEvent(eventID, command.TeacherID, command.FirstName, command.ChosenName, command.LastName, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateTeacherResult{}, err
	}
	return UpdateTeacherResult{TeacherUpdatedID: eventID}, nil
}

type updateTeacherContext struct {
	exists     bool
	deleted    bool
	firstName  string
	chosenName string
	lastName   string
	position   eventstore.Position
	events     []eventstore.ResolvedEvent
	query      eventstore.Query
}

func loadUpdateTeacherContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updateTeacherContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateTeacherContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateTeacherContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateTeacherContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case TeacherCreated:
		m.exists = true
		m.deleted = false
		m.firstName, _ = data["firstName"].(string)
		m.chosenName, _ = data["chosenName"].(string)
		m.lastName, _ = data["lastName"].(string)
	case TeacherUpdated:
		m.firstName, _ = data["firstName"].(string)
		m.chosenName, _ = data["chosenName"].(string)
		m.lastName, _ = data["lastName"].(string)
	case TeacherDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
