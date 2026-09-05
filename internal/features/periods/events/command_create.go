package events

import (
	"context"
	"time"

	"seek/internal/features/_shared/sharedmodels"
	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreatePeriodCommand struct {
	Title       string
	ServiceType sharedmodels.ServiceType
	StartTime   sharedmodels.TimeOnly
	Duration    int
	DaysBitmask sharedmodels.DaysBitmask
	Metadata    CommandMetadata
}

type CreatePeriodResult struct {
	EventID string
}

func CreatePeriodCommandHandler(
	ctx context.Context,
	cmd CreatePeriodCommand,
	saver eventstore.Saver,
) (
	CreatePeriodResult,
	error,
) {
	context, err := newCreatePeriodContext(cmd)
	if err != nil {
		return CreatePeriodResult{}, err
	}
	event := NewPeriodCreatedEvent(
		context.id,
		context.title,
		context.serviceType,
		context.startTime,
		context.duration,
		context.daysBitmask,
		time.Now(),
		metadataWithQuery(cmd.Metadata, context.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, context.query); err != nil {
		return CreatePeriodResult{}, err
	}
	return CreatePeriodResult{EventID: context.id}, nil
}

type createPeriodContext struct {
	id          string
	title       string
	serviceType sharedmodels.ServiceType
	startTime   sharedmodels.TimeOnly
	duration    int
	daysBitmask sharedmodels.DaysBitmask
	query       eventstore.Query
}

func newCreatePeriodContext(cmd CreatePeriodCommand) (*createPeriodContext, error) {
	periodID := uuidv7.NewString()
	return &createPeriodContext{
		id:          periodID,
		title:       cmd.Title,
		serviceType: cmd.ServiceType,
		startTime:   cmd.StartTime,
		duration:    cmd.Duration,
		daysBitmask: cmd.DaysBitmask,
		query:       streamQuery(periodID),
	}, nil
}
