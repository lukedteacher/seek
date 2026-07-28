package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type DeleteIEPServiceCommand struct {
	IEPServiceID string
	StudentID    string
	Metadata     CommandMetadata
}

type DeleteIEPServiceResult struct {
	EventID string
}

func DeleteIEPServiceCommandHandler(
	ctx context.Context,
	command DeleteIEPServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteIEPServiceResult,
	error,
) {
	model, err := loadDeleteIEPServiceContext(ctx, retriever, command.IEPServiceID, command.StudentID)
	if err != nil {
		return DeleteIEPServiceResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteIEPServiceResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewIEPServiceDeletedEvent(
		eventID,
		command.IEPServiceID,
		command.StudentID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteIEPServiceResult{}, err
	}
	return DeleteIEPServiceResult{EventID: eventID}, nil
}

type deleteIEPServiceContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteIEPServiceContext(ctx context.Context, retriever eventstore.Retriever, iepServiceID, studentID string) (*deleteIEPServiceContext, error) {
	query := streamQuery(iepServiceID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteIEPServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteIEPServiceContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *deleteIEPServiceContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventTypeIEPServiceAddedToStudent:
		m.exists = true
		m.deleted = false
	case EventTypeIEPServiceDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
