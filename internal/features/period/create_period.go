package period

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreatePeriodCommand struct {
	Title     string
	StartTime string
	Duration  int64
	Days      int64
	Metadata  CommandMetadata
}

type CreatePeriodResult struct {
	PeriodID string
}

func CreatePeriodCommandHandler(ctx context.Context, command CreatePeriodCommand, saver eventstore.Saver) (CreatePeriodResult, error) {
	context, err := newCreatePeriodContext(command)
	if err != nil {
		return CreatePeriodResult{}, err
	}
	event := NewPeriodCreatedEvent(
		context.id,
		context.title,
		context.startTime,
		context.duration,
		context.days,
		time.Now(),
		metadataWithQuery(command.Metadata, context.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, context.query); err != nil {
		return CreatePeriodResult{}, err
	}
	return CreatePeriodResult{PeriodID: context.id}, nil
}

type createPeriodContext struct {
	id        string
	title     string
	startTime string
	duration  int64
	days      int64
	query     eventstore.Query
}

func newCreatePeriodContext(command CreatePeriodCommand) (*createPeriodContext, error) {
	periodID := uuidv7.NewString()
	return &createPeriodContext{
		id:        periodID,
		title:     command.Title,
		startTime: command.StartTime,
		duration:  command.Duration,
		days:      command.Days,
		query:     streamQuery(periodID),
	}, nil
}
