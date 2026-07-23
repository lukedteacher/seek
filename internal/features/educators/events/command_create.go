package events

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreateEducatorCommand struct {
	GivenName  string
	ChosenName string
	FamilyName string
	Role       string
	Email      string
	Metadata   CommandMetadata
}

type CreateEducatorResult struct {
	EventID string
}

func CreateEducatorCommandHandler(
	ctx context.Context,
	command CreateEducatorCommand,
	saver eventstore.Saver,
) (
	CreateEducatorResult,
	error,
) {
	model, err := newCreateEducatorContext(command)
	if err != nil {
		return CreateEducatorResult{}, err
	}

	event := NewEducatorCreatedEvent(
		model.id,
		model.givenName,
		model.chosenName,
		model.familyName,
		model.role,
		model.email,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		eventstore.NoEventPosition,
		nil,
		model.query,
	); err != nil {
		return CreateEducatorResult{}, err
	}
	return CreateEducatorResult{EventID: model.id}, nil
}

type createEducatorContext struct {
	id         string
	givenName  string
	chosenName string
	familyName string
	role       string
	email      string
	query      eventstore.Query
}

func newCreateEducatorContext(command CreateEducatorCommand) (*createEducatorContext, error) {
	id := uuidv7.NewString()
	return &createEducatorContext{
		id:         id,
		givenName:  command.GivenName,
		chosenName: command.ChosenName,
		familyName: command.FamilyName,
		role:       command.Role,
		email:      command.Email,
		query:      streamQuery(id),
	}, nil
}
