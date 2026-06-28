package period

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type UpdatePeriodCommand struct {
	UserRegisteredID string
	Id               string
	StartTime        string
	Duration         int64
	Days             int64
	Title            string
	Metadata         CommandMetadata
}

type UpdatePeriodResult struct {
	PeriodUpdatedID string
	Skipped         bool
}

func UpdatePeriodCommandHandler(ctx context.Context, command UpdatePeriodCommand, saver eventstore.Saver, retriever eventstore.Retriever) (UpdatePeriodResult, error) {
	model, err := loadUpdatePeriodContext(ctx, retriever, command.Id)
	if err != nil {
		return UpdatePeriodResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return UpdatePeriodResult{}, err
	}
	if model.title == command.Title && model.startTime == command.StartTime && model.duration == command.Duration && model.days == command.Days {
		return UpdatePeriodResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewPeriodUpdatedEvent(eventID, command.Id, command.Title, command.StartTime, command.Duration, command.Days, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdatePeriodResult{}, err
	}
	return UpdatePeriodResult{PeriodUpdatedID: eventID}, nil
}

type updatePeriodContext struct {
	exists    bool
	deleted   bool
	title     string
	startTime string
	duration  int64
	days      int64
	position  eventstore.Position
	events    []eventstore.ResolvedEvent
	query     eventstore.Query
}

func loadUpdatePeriodContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updatePeriodContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updatePeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updatePeriodContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updatePeriodContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case PeriodCreated:
		m.exists = true
		m.deleted = false
		m.title, _ = data["title"].(string)
		m.startTime, _ = data["startTime"].(string)
		m.duration = int64(data["duration"].(float64))
		m.days = int64(data["days"].(float64))
	case PeriodUpdated:
		m.title, _ = data["title"].(string)
		m.startTime, _ = data["startTime"].(string)
		m.duration = int64(data["duration"].(float64))
		m.days = int64(data["days"].(float64))
	case PeriodDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
