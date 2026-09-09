package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
)

type DeleteHomeroomCommand struct {
	HomeroomID string
	Metadata   CommandMetadata
}

type DeleteHomeroomResult struct {
	HomeroomDeletedID string
}

func DeleteHomeroomCommandHandler(
	ctx context.Context,
	cmd DeleteHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteHomeroomResult,
	error,
) {
	model, err := loadDeleteHomeroomContext(ctx, retriever, cmd.HomeroomID)
	if err != nil {
		return DeleteHomeroomResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteHomeroomResult{}, err
	}
	event := NewHomeroomDeletedEvent(
		cmd.HomeroomID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteHomeroomResult{}, err
	}
	return DeleteHomeroomResult{HomeroomDeletedID: event.EventID}, nil
}

type deleteHomeroomContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteHomeroomContext(ctx context.Context, retriever eventstore.Retriever, homeroomID string) (*deleteHomeroomContext, error) {
	query := homeroomStreamQuery(homeroomID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteHomeroomContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteHomeroomContext) requireActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func (m *deleteHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.created = true
	case EventHomeroomArchived:
		m.archived = true
	case EventHomeroomDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
