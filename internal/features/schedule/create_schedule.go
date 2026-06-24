package schedule

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreateScheduleCommand struct {
	Title     string
	TeacherId string
	Metadata  CommandMetadata
}

type CreateScheduleResult struct {
	Id string
}

func CreateScheduleCommandHandler(ctx context.Context, command CreateScheduleCommand, saver eventstore.Saver) (CreateScheduleResult, error) {
	model, err := newCreateScheduleContext(command)
	if err != nil {
		return CreateScheduleResult{}, err
	}

	event := NewScheduleCreatedEvent(model.id, model.title, model.teacherId, time.Now(), metadataWithQuery(command.Metadata, model.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateScheduleResult{}, err
	}
	return CreateScheduleResult{Id: model.id}, nil
}

type createScheduleContext struct {
	id        string
	title     string
	teacherId string
	query     eventstore.Query
}

func newCreateScheduleContext(command CreateScheduleCommand) (*createScheduleContext, error) {
	id := uuidv7.NewString()
	return &createScheduleContext{
		id:        id,
		title:     command.Title,
		teacherId: command.TeacherId,
		query:     streamQuery(id),
	}, nil
}
