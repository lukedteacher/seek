package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type UpdateTeacherCommand struct {
	UserRegisteredID string
	TeacherID        string
	GivenName        string
	ChosenName       string
	FamilyName       string
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
	if model.givenName == command.GivenName && model.chosenName == command.ChosenName && model.familyName == command.FamilyName {
		return UpdateTeacherResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewTeacherUpdatedEvent(eventID, command.TeacherID, command.GivenName, command.ChosenName, command.FamilyName, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateTeacherResult{}, err
	}
	return UpdateTeacherResult{TeacherUpdatedID: eventID}, nil
}

type updateTeacherContext struct {
	exists     bool
	deleted    bool
	givenName  string
	chosenName string
	familyName string
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
		m.givenName, _ = data[TeacherGivenNameField].(string)
		m.chosenName, _ = data[TeacherChosenNameField].(string)
		m.familyName, _ = data[TeacherFamilyNameField].(string)
	case TeacherUpdated:
		m.givenName, _ = data[TeacherGivenNameField].(string)
		m.chosenName, _ = data[TeacherChosenNameField].(string)
		m.familyName, _ = data[TeacherFamilyNameField].(string)
	case TeacherDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
