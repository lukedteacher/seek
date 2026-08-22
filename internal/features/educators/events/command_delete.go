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
	if !model.isActive() {
		return DeleteEducatorResult{}, eventstore.ErrEducatorNotActive
	}
	eventID := uuidv7.NewString()

	// build event data struct directly
	eventData := EducatorDeletedEvent{
		EventID:   eventID,
		DeletedAt: time.Now(),
		Scope:     educatorScope(command.EducatorID),
	}

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorDeleted,
		Data:      eventstore.MustData(eventData),
		Metadata:  metadataWithQuery(command.Metadata, model.query),
	}

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
	created  bool
	archived bool
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

func (m *deleteEducatorContext) isActive() bool {
	if !m.created || m.archived || m.deleted {
		return false
	}
	return true
}

func (m *deleteEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventEducatorCreated:
		m.created = true
	case EventEducatorArchived:
		m.archived = true
	case EventEducatorDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
