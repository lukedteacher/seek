package schedule

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type RemovePeriodFromScheduleCommand struct {
	EventID    string
	ScheduleID string
	PeriodID   string
	Metadata   CommandMetadata
}

type SchedulePeriodRemoveResult struct {
	EventID string
	Skipped bool
}

func RemovePeriodFromScheduleCommandHandler(ctx context.Context, command RemovePeriodFromScheduleCommand, saver eventstore.Saver, retriever eventstore.Retriever) (*SchedulePeriodRemoveResult, error) {
	model, err := loadRemovePeriodFromScheduleContext(ctx, retriever, command.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := model.requireActive(); err != nil {
		return nil, err
	}

	skipped := !model.added
	if skipped {
		return &SchedulePeriodRemoveResult{Skipped: skipped}, nil
	}
	eventID := uuidv7.NewString()
	event := NewSchedulePeriodRemovedEvent(eventID, command.ScheduleID, command.PeriodID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &SchedulePeriodRemoveResult{EventID: eventID, Skipped: false}, nil
}

type removePeriodFromScheduleContext struct {
	exists   bool
	deleted  bool
	added    bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadRemovePeriodFromScheduleContext(ctx context.Context, retriever eventstore.Retriever, id string) (*removePeriodFromScheduleContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &removePeriodFromScheduleContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *removePeriodFromScheduleContext) requireActive() error {
	if !c.exists || c.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (c *removePeriodFromScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case ScheduleCreated:
		c.exists = true
		c.deleted = false
	case ScheduleDeleted:
		c.deleted = true
	case SchedulePeriodAdded:
		c.added = true
	case SchedulePeriodRemoved:
		c.added = false
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
