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

type PeriodAddedToScheduleResult struct {
	EventID string
	Skipped bool
}

func PeriodAddedToScheduleCommandHandler(ctx context.Context, command PeriodAddedToScheduleCommand, saver eventstore.Saver, retriever eventstore.Retriever) (*PeriodAddedToScheduleResult, error) {
	model, err := loadPeriodAddedToScheduleContext(ctx, retriever, command.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := model.requireActive(); err != nil {
		return nil, err
	}
	// TODO find out how to compare slices
	// if model.periodIDs == command.PeriodIDs {
	// 	return &PeriodAddedToScheduleResult{Skipped: true}, nil
	// }
	eventID := uuidv7.NewString()
	event := NewPeriodAddedToScheduleEvent(eventID, command.ScheduleID, command.PeriodID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &PeriodAddedToScheduleResult{EventID: eventID, Skipped: false}, nil
}

type periodAddedToScheduleContext struct {
	exists   bool
	deleted  bool
	periodID string
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

func (m *periodAddedToScheduleContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *periodAddedToScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case ScheduleCreated:
		m.exists = true
		m.deleted = false
	case PeriodAddedToSchedule:
		// TODO should there be more to this?
		m.periodID = data["periodID"].(string)
	case ScheduleDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
