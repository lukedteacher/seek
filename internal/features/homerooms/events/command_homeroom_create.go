package events

import (
	"context"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
	"seek/internal/features/homerooms/models"
)

type CreateHomeroomCommand struct {
	Homeroom models.Homeroom
	Metadata CommandMetadata
}

type CreateHomeroomResult struct {
	EventID string
}

func CreateHomeroomCommandHandler(
	ctx context.Context,
	cmd CreateHomeroomCommand,
	saver eventstore.Saver,
) (
	CreateHomeroomResult,
	error,
) {
	context, err := newCreateHomeroomContext(cmd)
	if err != nil {
		return CreateHomeroomResult{}, err
	}
	event := NewHomeroomCreatedEvent(
		context.id,
		context.homeroom,
		time.Now(),
		metadataWithQuery(cmd.Metadata, context.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, context.query); err != nil {
		return CreateHomeroomResult{}, err
	}
	return CreateHomeroomResult{EventID: context.id}, nil
}

type createHomeroomContext struct {
	id       string
	homeroom models.Homeroom
	query    eventstore.Query
}

func newCreateHomeroomContext(cmd CreateHomeroomCommand) (*createHomeroomContext, error) {
	homeroomID := uuidv7.NewString()
	return &createHomeroomContext{
		id:       homeroomID,
		homeroom: cmd.Homeroom,
		query:    homeroomStreamQuery(homeroomID),
	}, nil
}
