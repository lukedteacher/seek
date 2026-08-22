package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type ArchivePeriodCommand struct {
	PeriodID string
	Metadata CommandMetadata
}

type ArchivePeriodResult struct {
	PeriodArchivedID string
}

func ArchivePeriodCommandHandler(
	ctx context.Context,
	command ArchivePeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	ArchivePeriodResult,
	error,
) {
	model, err := loadArchivePeriodContext(ctx, retriever, command.PeriodID)
	if err != nil {
		return ArchivePeriodResult{}, err
	}
	if err := model.isActive(); err != nil {
		return ArchivePeriodResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewPeriodArchivedEvent(eventID, command.PeriodID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return ArchivePeriodResult{}, err
	}
	return ArchivePeriodResult{PeriodArchivedID: eventID}, nil
}

type archivePeriodContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadArchivePeriodContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
) (
	*archivePeriodContext,
	error,
) {
	query := streamQuery(periodID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &archivePeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *archivePeriodContext) isActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func (m *archivePeriodContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventPeriodCreated:
		m.created = true
	case EventPeriodArchived:
		m.archived = true
	case EventPeriodDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
