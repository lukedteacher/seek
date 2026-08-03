package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
	"seek/pkg/uuidv7"
)

type AddEducatorToPeriodCommand struct {
	PeriodID   string
	EducatorID string
	Metadata   CommandMetadata
}

type AddEducatorToPeriodResult struct {
	EventID string
	Skipped bool
}

func AddEducatorToPeriodCommandHandler(
	ctx context.Context,
	command AddEducatorToPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddEducatorToPeriodResult,
	error,
) {
	model, err := loadAddEducatorToPeriodContext(ctx, retriever, command.PeriodID, command.EducatorID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isEducatorActive(); err != nil {
		return nil, err
	}

	skip := model.added
	if skip {
		return &AddEducatorToPeriodResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewEducatorAddedToPeriodEvent(
		eventID,
		command.PeriodID,
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
		return nil, err
	}
	return &AddEducatorToPeriodResult{EventID: eventID, Skipped: false}, nil
}

type addEducatorToPeriodContext struct {
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

func loadAddEducatorToPeriodContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	educatorID string,
) (
	*addEducatorToPeriodContext,
	error,
) {
	query := educatorPeriodStreamQuery(periodID, educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &addEducatorToPeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *addEducatorToPeriodContext) isPeriodActive() error {
	if !m.periodCreated || m.periodArchived || m.periodDeleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func (m *addEducatorToPeriodContext) isEducatorActive() error {
	if !m.educatorCreated || m.educatorArchived || m.educatorDeleted {
		return eventstore.ErrEducatorNotActive
	}
	return nil
}

func (m *addEducatorToPeriodContext) handle(resolved eventstore.ResolvedEvent) {
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
