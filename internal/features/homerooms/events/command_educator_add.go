package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	educatorEvents "seek/internal/features/educators/events"
)

type AddEducatorToHomeroomCommand struct {
	HomeroomID string
	EducatorID string
	Metadata   CommandMetadata
}

type AddEducatorToHomeroomResult struct {
	EventID string
	Skipped bool
}

func AddEducatorToHomeroomCommandHandler(
	ctx context.Context,
	cmd AddEducatorToHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddEducatorToHomeroomResult,
	error,
) {
	model, err := loadAddEducatorToHomeroomContext(ctx, retriever, cmd.HomeroomID, cmd.EducatorID)
	if err != nil {
		return nil, err
	}
	if err := model.isHomeroomActive(); err != nil {
		return nil, err
	}
	if err := model.isEducatorActive(); err != nil {
		return nil, err
	}
	skip := model.educator.added
	if skip {
		return &AddEducatorToHomeroomResult{Skipped: skip}, nil
	}

	event := NewEducatorAddedToHomeroomEvent(
		cmd.HomeroomID,
		cmd.EducatorID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
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
	return &AddEducatorToHomeroomResult{EventID: event.EventID, Skipped: false}, nil
}

type homeroomState struct {
	created  bool
	archived bool
	deleted  bool
}

type educatorState struct {
	created  bool
	archived bool
	deleted  bool
	added    bool
}

type addEducatorToHomeroomContext struct {
	homeroom homeroomState
	educator educatorState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadAddEducatorToHomeroomContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	homeroomID,
	educatorID string,
) (
	*addEducatorToHomeroomContext,
	error,
) {
	homeroomEducatorQuery := homeroomEducatorStreamQuery(homeroomID, educatorID)
	educatorQuery := educatorEvents.StreamQuery(educatorID)
	query := combineQueries(homeroomEducatorQuery, educatorQuery)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &addEducatorToHomeroomContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *addEducatorToHomeroomContext) isHomeroomActive() error {
	if !m.homeroom.created || m.homeroom.archived || m.homeroom.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func (m *addEducatorToHomeroomContext) isEducatorActive() error {
	if !m.educator.created || m.educator.archived || m.educator.deleted {
		return eventstore.ErrEducatorNotActive
	}
	return nil
}

func (m *addEducatorToHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.homeroom.created = true
	case EventHomeroomArchived:
		m.homeroom.archived = true
	case EventHomeroomDeleted:
		m.homeroom.deleted = true
	case educatorEvents.EventEducatorCreated:
		m.educator.created = true
	case educatorEvents.EventEducatorArchived:
		m.educator.archived = true
	case educatorEvents.EventEducatorDeleted:
		m.educator.deleted = true
	case EventEducatorAddedToHomeroom:
		m.educator.added = true
	case EventEducatorRemovedFromHomeroom:
		m.educator.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
