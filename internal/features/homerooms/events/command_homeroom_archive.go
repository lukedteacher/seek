package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type ArchiveHomeroomCommand struct {
	HomeroomID string
	Metadata   CommandMetadata
}

type ArchiveHomeroomResult struct {
	HomeroomArchivedID string
}

func ArchiveHomeroomCommandHandler(
	ctx context.Context,
	cmd ArchiveHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	ArchiveHomeroomResult,
	error,
) {
	model, err := loadArchiveHomeroomContext(ctx, retriever, cmd.HomeroomID)
	if err != nil {
		return ArchiveHomeroomResult{}, err
	}
	if err := model.isActive(); err != nil {
		return ArchiveHomeroomResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewHomeroomArchivedEvent(
		cmd.HomeroomID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return ArchiveHomeroomResult{}, err
	}
	return ArchiveHomeroomResult{HomeroomArchivedID: eventID}, nil
}

type archiveHomeroomContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadArchiveHomeroomContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	homeroomID string,
) (
	*archiveHomeroomContext,
	error,
) {
	query := homeroomStreamQuery(homeroomID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &archiveHomeroomContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *archiveHomeroomContext) isActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func (m *archiveHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
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
