package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type ArchiveEducatorCommand struct {
	EducatorID string
	Metadata   CommandMetadata
}

type ArchiveEducatorResult struct {
	EventID string
}

func ArchiveEducatorCommandHandler(
	ctx context.Context,
	command ArchiveEducatorCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	ArchiveEducatorResult,
	error,
) {
	model, err := loadArchiveEducatorContext(ctx, retriever, command.EducatorID)
	if err != nil {
		return ArchiveEducatorResult{}, err
	}
	if !model.isActive() {
		return ArchiveEducatorResult{}, eventstore.ErrEducatorNotActive
	}

	eventID := uuidv7.NewString()

	// build event data struct directly
	eventData := EducatorArchivedEvent{
		EventID:    eventID,
		ArchivedAt: time.Now(),
		Scope:      educatorScope(command.EducatorID),
	}

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorArchived,
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
		return ArchiveEducatorResult{}, err
	}
	return ArchiveEducatorResult{EventID: eventID}, nil
}

type archiveEducatorContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadArchiveEducatorContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	educatorID string,
) (
	*archiveEducatorContext,
	error,
) {
	query := streamQuery(educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &archiveEducatorContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *archiveEducatorContext) isActive() bool {
	if !m.created || m.archived || m.deleted {
		return false
	}
	return true
}

func (m *archiveEducatorContext) handle(resolved eventstore.ResolvedEvent) {
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
