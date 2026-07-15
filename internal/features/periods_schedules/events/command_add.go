package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/schedules/events"
	"seek/internal/uuidv7"
)

type PeriodScheduleAddCommand struct {
	EventID    string
	PeriodID   string
	ScheduleID string
	Metadata   CommandMetadata
}

type PeriodScheduleAddResult struct {
	EventID string
	Skipped bool
}

func PeriodScheduleAddCommandHandler(
	ctx context.Context,
	command PeriodScheduleAddCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*PeriodScheduleAddResult,
	error,
) {
	model, err := loadPeriodScheduleAddContext(ctx, retriever, command.PeriodID, command.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isScheduleActive(); err != nil {
		return nil, err
	}

	skip := model.added
	if skip {
		return &PeriodScheduleAddResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewPeriodScheduleAddedEvent(
		eventID,
		command.PeriodID,
		command.ScheduleID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &PeriodScheduleAddResult{EventID: eventID, Skipped: false}, nil
}

type periodScheduleAddContext struct {
	periodCreated   bool
	periodDeleted   bool
	scheduleCreated bool
	scheduleDeleted bool
	added           bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadPeriodScheduleAddContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	scheduleID string,
) (
	*periodScheduleAddContext,
	error,
) {
	query := streamQuery(periodID, scheduleID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &periodScheduleAddContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *periodScheduleAddContext) isPeriodActive() error {
	if !c.periodCreated || c.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (c *periodScheduleAddContext) isScheduleActive() error {
	if !c.scheduleCreated || c.scheduleDeleted {
		return eventstore.ErrScheduleNotFound
	}
	return nil
}

func (c *periodScheduleAddContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		c.periodCreated = true
		c.periodDeleted = false
	case pe.PeriodDeleted:
		c.periodDeleted = true
	case se.ScheduleCreated:
		c.scheduleCreated = true
		c.scheduleDeleted = false
	case se.ScheduleDeleted:
		c.scheduleDeleted = true
	case PeriodScheduleAdded:
		c.added = true
	case PeriodScheduleRemoved:
		c.added = false
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
