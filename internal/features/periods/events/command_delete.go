package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type DeletePeriodCommand struct {
	PeriodID string
	Metadata CommandMetadata
}

type DeletePeriodResult struct {
	PeriodDeletedID string
}

func DeletePeriodCommandHandler(
	ctx context.Context,
	command DeletePeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeletePeriodResult,
	error,
) {
	model, err := loadDeletePeriodContext(ctx, retriever, command.PeriodID)
	if err != nil {
		return DeletePeriodResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeletePeriodResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewPeriodDeletedEvent(eventID, command.PeriodID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeletePeriodResult{}, err
	}
	return DeletePeriodResult{PeriodDeletedID: eventID}, nil
}

type deletePeriodContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeletePeriodContext(ctx context.Context, retriever eventstore.Retriever, periodID string) (*deletePeriodContext, error) {
	query := streamQuery(periodID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deletePeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deletePeriodContext) requireActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func (m *deletePeriodContext) handle(resolved eventstore.ResolvedEvent) {
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
