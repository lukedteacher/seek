package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type DeleteEducatorCommand struct {
	EducatorID string
	Metadata   CommandMetadata
}

type DeleteEducatorResult struct {
	EventID string
}

func DeleteEducatorCommandHandler(
	ctx context.Context,
	command DeleteEducatorCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteEducatorResult,
	error,
) {
	model, err := loadDeleteEducatorContext(ctx, retriever, command.EducatorID)
	if err != nil {
		return DeleteEducatorResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteEducatorResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewEducatorDeletedEvent(
		eventID,
		command.EducatorID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		model.events,
		model.query,
	); err != nil {
		return DeleteEducatorResult{}, err
	}
	return DeleteEducatorResult{EventID: eventID}, nil
}

type deleteEducatorContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteEducatorContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	educatorID string,
) (
	*deleteEducatorContext,
	error,
) {
	query := streamQuery(educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteEducatorContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteEducatorContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *deleteEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EducatorCreated:
		m.exists = true
		m.deleted = false
	case EducatorDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
