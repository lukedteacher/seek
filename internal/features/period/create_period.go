package period

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreatePeriodCommand struct {
	Title string
	StartTime string
	Duration int64
	Days int64
	Metadata         CommandMetadata
}

type CreatePeriodResult struct {
	Id string
}

func CreatePeriodCommandHandler(ctx context.Context, command CreatePeriodCommand, saver eventstore.Saver) (CreatePeriodResult, error) {
	model, err := newCreatePeriodContext(command)
	if err != nil {
		return CreatePeriodResult{}, err
	}

	event := NewPeriodCreatedEvent(model.id, model.title, model.startTime, model.duration, model.days, time.Now(), metadataWithQuery(command.Metadata, model.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreatePeriodResult{}, err
	}
	return CreatePeriodResult{Id: model.id}, nil
}

type createPeriodContext struct {
	id string
	title  string
	startTime string
	duration int64
	days int64
	query  eventstore.Query
}

func newCreatePeriodContext(command CreatePeriodCommand) (*createPeriodContext, error) {
	// firstName, err := validateTitle(command.FirstName)
	// if err != nil {
	// 	return nil, err
	// }
	id := uuidv7.NewString()
	return &createPeriodContext{
		id: id,
		title:  command.Title,
		startTime:  command.StartTime,
		duration:  command.Duration,
		days:  command.Days,
		query:  streamQuery(id),
	}, nil
}
