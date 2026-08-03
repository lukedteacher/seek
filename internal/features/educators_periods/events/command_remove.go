package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
	"seek/pkg/uuidv7"
)

type RemoveEducatorFromPeriodCommand struct {
	PeriodID   string
	EducatorID string
	Metadata   CommandMetadata
}

type RemoveEducatorFromPeriodResult struct {
	EventID string
	Skipped bool
}

func RemoveEducatorFromPeriodCommandHandler(
	ctx context.Context,
	command RemoveEducatorFromPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*RemoveEducatorFromPeriodResult,
	error,
) {
	model, err := loadRemoveEducatorFromPeriodContext(ctx, retriever, command.PeriodID, command.EducatorID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isEducatorActive(); err != nil {
		return nil, err
	}

	skip := !model.added
	if skip {
		return &RemoveEducatorFromPeriodResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewEducatorRemovedFromPeriodEvent(eventID, command.PeriodID, command.EducatorID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &RemoveEducatorFromPeriodResult{EventID: eventID, Skipped: false}, nil
}

type removeEducatorFromPeriodContext struct {
	periodCreated    bool
	periodArchived   bool
	periodDeleted    bool
	educatorCreated  bool
	educatorArchived bool
	educatorDeleted  bool
	added            bool
	position         eventstore.Position
	events           []eventstore.ResolvedEvent
	query            eventstore.Query
}

func loadRemoveEducatorFromPeriodContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	educatorID string,
) (
	*removeEducatorFromPeriodContext,
	error,
) {
	query := educatorPeriodStreamQuery(periodID, educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &removeEducatorFromPeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *removeEducatorFromPeriodContext) isPeriodActive() error {
	if !m.periodCreated || m.periodArchived || m.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *removeEducatorFromPeriodContext) isEducatorActive() error {
	if !m.educatorCreated || m.educatorArchived || m.educatorDeleted {
		return eventstore.ErrEducatorNotFound
	}
	return nil
}

func (m *removeEducatorFromPeriodContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		m.periodCreated = true
	case pe.PeriodArchived:
		m.periodArchived = true
	case pe.PeriodDeleted:
		m.periodDeleted = true
	case se.EducatorCreated:
		m.educatorCreated = true
	case se.EducatorArchived:
		m.educatorArchived = true
	case se.EducatorDeleted:
		m.educatorDeleted = true
	case EducatorAddedToPeriod:
		m.added = true
	case EducatorRemovedFromPeriod:
		m.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
