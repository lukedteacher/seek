package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type DeleteScheduleCommand struct {
	ScheduleID string
	Metadata   CommandMetadata
}

type DeleteScheduleResult struct {
	ScheduleDeletedID string
}

func DeleteScheduleCommandHandler(ctx context.Context, command DeleteScheduleCommand, saver eventstore.Saver, retriever eventstore.Retriever) (DeleteScheduleResult, error) {
	model, err := loadDeleteScheduleContext(ctx, retriever, command.ScheduleID)
	if err != nil {
		return DeleteScheduleResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return DeleteScheduleResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewScheduleDeletedEvent(eventID, command.ScheduleID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteScheduleResult{}, err
	}
	return DeleteScheduleResult{ScheduleDeletedID: eventID}, nil
}

type deleteScheduleContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteScheduleContext(ctx context.Context, retriever eventstore.Retriever, scheduleID string) (*deleteScheduleContext, error) {
	query := streamQuery(scheduleID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteScheduleContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteScheduleContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *deleteScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case ScheduleCreated:
		m.exists = true
		m.deleted = false
	case ScheduleDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
