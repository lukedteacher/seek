package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type DeleteStudentServiceCommand struct {
	ServiceID string
	StudentID string
	Metadata  CommandMetadata
}

type DeleteStudentServiceResult struct {
	StudentServiceDeletedID string
}

func DeleteStudentServiceCommandHandler(
	ctx context.Context,
	command DeleteStudentServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteStudentServiceResult,
	error,
) {
	model, err := loadDeleteStudentServiceContext(ctx, retriever, command.ServiceID, command.StudentID)
	if err != nil {
		return DeleteStudentServiceResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteStudentServiceResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewStudentServiceDeletedEvent(
		eventID,
		command.ServiceID,
		command.StudentID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteStudentServiceResult{}, err
	}
	return DeleteStudentServiceResult{StudentServiceDeletedID: eventID}, nil
}

type deleteStudentServiceContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteStudentServiceContext(ctx context.Context, retriever eventstore.Retriever, serviceID, studentID string) (*deleteStudentServiceContext, error) {
	query := streamQuery(serviceID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteStudentServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteStudentServiceContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *deleteStudentServiceContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case StudentServiceCreated:
		m.exists = true
		m.deleted = false
	case StudentServiceDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
