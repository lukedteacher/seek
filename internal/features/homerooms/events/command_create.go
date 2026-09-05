package events

import (
	"context"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreateHomeroomCommand struct {
	Title      string
	LocationID string
	Metadata   CommandMetadata
}

type CreateHomeroomResult struct {
	EventID string
}

func CreateHomeroomCommandHandler(
	ctx context.Context,
	command CreateHomeroomCommand,
	saver eventstore.Saver,
) (
	CreateHomeroomResult,
	error,
) {
	context, err := newCreateHomeroomContext(command)
	if err != nil {
		return CreateHomeroomResult{}, err
	}
	event := NewHomeroomCreatedEvent(
		context.id,
		context.title,
		context.locationID,
		time.Now(),
		metadataWithQuery(command.Metadata, context.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, context.query); err != nil {
		return CreateHomeroomResult{}, err
	}
	return CreateHomeroomResult{EventID: context.id}, nil
}

type createHomeroomContext struct {
	id         string
	title      string
	locationID string
	query      eventstore.Query
}

func newCreateHomeroomContext(command CreateHomeroomCommand) (*createHomeroomContext, error) {
	homeroomID := uuidv7.NewString()
	return &createHomeroomContext{
		id:         homeroomID,
		title:      command.Title,
		locationID: command.LocationID,
		query:      homeroomStreamQuery(homeroomID),
	}, nil
}
