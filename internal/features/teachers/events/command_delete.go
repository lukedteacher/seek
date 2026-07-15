package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type DeleteTeacherCommand struct {
	TeacherID string
	Metadata  CommandMetadata
}

type DeleteTeacherResult struct {
	TeacherDeletedID string
}

func DeleteTeacherCommandHandler(ctx context.Context, command DeleteTeacherCommand, saver eventstore.Saver, retriever eventstore.Retriever) (DeleteTeacherResult, error) {
	model, err := loadDeleteTeacherContext(ctx, retriever, command.TeacherID)
	if err != nil {
		return DeleteTeacherResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteTeacherResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewTeacherDeletedEvent(eventID, command.TeacherID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteTeacherResult{}, err
	}
	return DeleteTeacherResult{TeacherDeletedID: eventID}, nil
}

type deleteTeacherContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteTeacherContext(ctx context.Context, retriever eventstore.Retriever, studentID string) (*deleteTeacherContext, error) {
	query := streamQuery(studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteTeacherContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteTeacherContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *deleteTeacherContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case TeacherCreated:
		m.exists = true
		m.deleted = false
	case TeacherDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
