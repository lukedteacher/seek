package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/schedules/events"
	"seek/internal/uuidv7"
)

type PeriodScheduleRemoveCommand struct {
	EventID   string
	PeriodID  string
	ScheduleID string
	Metadata  CommandMetadata
}

type PeriodScheduleRemoveResult struct {
	EventID string
	Skipped bool
}

func PeriodScheduleRemoveCommandHandler(
	ctx context.Context,
	command PeriodScheduleRemoveCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*PeriodScheduleRemoveResult,
	error,
) {
	model, err := loadPeriodScheduleRemoveContext(ctx, retriever, command.PeriodID, command.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isScheduleActive(); err != nil {
		return nil, err
	}

	skip := !model.scheduleAdded
	if skip {
		return &PeriodScheduleRemoveResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewPeriodScheduleRemovedEvent(eventID, command.PeriodID, command.ScheduleID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &PeriodScheduleRemoveResult{EventID: eventID, Skipped: false}, nil
}

type periodScheduleRemoveContext struct {
	periodCreated  bool
	periodDeleted  bool
	scheduleCreated bool
	scheduleDeleted bool
	scheduleAdded   bool
	position       eventstore.Position
	events         []eventstore.ResolvedEvent
	query          eventstore.Query
}

func loadPeriodScheduleRemoveContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	scheduleID string,
) (
	*periodScheduleRemoveContext,
	error,
) {
	query := streamQuery(periodID, scheduleID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &periodScheduleRemoveContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *periodScheduleRemoveContext) isPeriodActive() error {
	if !c.periodCreated || c.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (c *periodScheduleRemoveContext) isScheduleActive() error {
	if !c.scheduleCreated || c.scheduleDeleted {
		return eventstore.ErrScheduleNotFound
	}
	return nil
}

func (c *periodScheduleRemoveContext) handle(resolved eventstore.ResolvedEvent) {
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
		c.scheduleAdded = true
	case PeriodScheduleRemoved:
		c.scheduleAdded = false
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
