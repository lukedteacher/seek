package schedule

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type PeriodAddedToScheduleCommand struct {
	EventID    string
	ScheduleID string
	PeriodID   string
	Metadata   CommandMetadata
}

type SchedulePeriodAddedResult struct {
	EventID string
	Skipped bool
}

func PeriodAddedToScheduleCommandHandler(ctx context.Context, command PeriodAddedToScheduleCommand, saver eventstore.Saver, retriever eventstore.Retriever) (*SchedulePeriodAddedResult, error) {
	model, err := loadPeriodAddedToScheduleContext(ctx, retriever, command.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := model.requireActive(); err != nil {
		return nil, err
	}

	skipped := model.added
	if skipped {
		return &SchedulePeriodAddedResult{Skipped: skipped}, nil
	}
	eventID := uuidv7.NewString()
	event := NewSchedulePeriodAddedEvent(eventID, command.ScheduleID, command.PeriodID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &SchedulePeriodAddedResult{EventID: eventID, Skipped: false}, nil
}

type periodAddedToScheduleContext struct {
	exists   bool
	deleted  bool
	added    bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadPeriodAddedToScheduleContext(ctx context.Context, retriever eventstore.Retriever, id string) (*periodAddedToScheduleContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &periodAddedToScheduleContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *periodAddedToScheduleContext) requireActive() error {
	if !c.exists || c.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (c *periodAddedToScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case ScheduleCreated:
		c.exists = true
		c.deleted = false
	case SchedulePeriodAdded:
		c.added = true
	case SchedulePeriodRemoved:
		c.added = false
	case ScheduleDeleted:
		c.deleted = true
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
