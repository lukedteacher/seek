package schedule

import (
	"context"
	"slices"
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

	skipped := true
	if len(model.periodIDs) > 0 {
		for _, periodID := range model.periodIDs {
			if periodID == command.PeriodID {
				skipped = false
			}
		}
	}
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
	exists    bool
	deleted   bool
	periodIDs []string
	position  eventstore.Position
	events    []eventstore.ResolvedEvent
	query     eventstore.Query
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

func (m *removePeriodFromScheduleContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *removePeriodFromScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case ScheduleCreated:
		m.exists = true
		m.deleted = false
	case SchedulePeriodAdded:
		// TODO should there be more to this?
		m.periodIDs = append(m.periodIDs, data["periodID"].(string))
	case SchedulePeriodRemoved:
		periodToRemove := data["periodID"].(string)
		m.periodIDs = slices.DeleteFunc(m.periodIDs, func(id string) bool {
        return id == periodToRemove
    })
	case ScheduleDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
