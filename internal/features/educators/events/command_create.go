package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"
	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreateEducatorCommand struct {
	Person   sharedmodels.Person
	Role     string
	Metadata CommandMetadata
}

type CreateEducatorResult struct {
	EventID string
}

func CreateEducatorCommandHandler(
	ctx context.Context,
	command CreateEducatorCommand,
	saver eventstore.Saver,
) (CreateEducatorResult, error) {
	model, err := newCreateEducatorContext(command)
	if err != nil {
		return CreateEducatorResult{}, fmt.Errorf("building educator context: %w", err)
	}

	event := NewEducatorCreatedEvent(
		model.id,
		model.givenName,
		model.chosenName,
		model.familyName,
		model.email,
		model.role,
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
		return CreateEducatorResult{}, fmt.Errorf("saving educator created event: %w", err)
	}

	return CreateEducatorResult{EventID: model.id}, nil
}

type createEducatorContext struct {
	id         string
	givenName  string
	chosenName string
	familyName string
	email      string
	role       string
	query      eventstore.Query
}

func newCreateEducatorContext(command CreateEducatorCommand) (*createEducatorContext, error) {
	id := uuidv7.NewString()

	educator, err := models.NewEducator(
		id,
		command.Person.GivenName,
		command.Person.ChosenName,
		command.Person.FamilyName,
		command.Person.Email,
		command.Role,
	)
	if err != nil {
		return nil, err
	}

	return &createEducatorContext{
		id:         id,
		givenName:  educator.GivenName,
		chosenName: educator.ChosenName,
		familyName: educator.FamilyName,
		email:      educator.Email,
		role:       educator.Role,
		query:      streamQuery(id),
	}, nil
}
