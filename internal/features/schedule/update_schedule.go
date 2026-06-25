package schedule

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type UpdateScheduleCommand struct {
	UserRegisteredID string
	Id               string
	Title            string
	TeacherId        string
	Metadata         CommandMetadata
}

type UpdateScheduleResult struct {
	ScheduleUpdatedID string
	Skipped           bool
}

func UpdateScheduleCommandHandler(ctx context.Context, command UpdateScheduleCommand, saver eventstore.Saver, retriever eventstore.Retriever) (UpdateScheduleResult, error) {
	model, err := loadUpdateScheduleContext(ctx, retriever, command.Id)
	if err != nil {
		return UpdateScheduleResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return UpdateScheduleResult{}, err
	}
	if model.title == command.Title && model.teacherID == command.TeacherId {
		return UpdateScheduleResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewScheduleUpdatedEvent(eventID, command.Id, command.Title, command.TeacherId, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateScheduleResult{}, err
	}
	return UpdateScheduleResult{ScheduleUpdatedID: eventID}, nil
}

type updateScheduleContext struct {
	exists    bool
	deleted   bool
	title     string
	teacherID string
	position  eventstore.Position
	events    []eventstore.ResolvedEvent
	query     eventstore.Query
}

func loadUpdateScheduleContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updateScheduleContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateScheduleContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateScheduleContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateScheduleContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case ScheduleCreated:
		m.exists = true
		m.deleted = false
		m.title, _ = data["title"].(string)
		m.teacherID, _ = data["start_time"].(string)
	case ScheduleUpdated:
		m.title, _ = data["title"].(string)
		m.teacherID, _ = data["start_time"].(string)
	case ScheduleDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
