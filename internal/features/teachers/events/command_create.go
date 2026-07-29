package events

import (
	"context"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreateTeacherCommand struct {
	GivenName  string
	ChosenName string
	FamilyName string
	Metadata   CommandMetadata
}

type CreateTeacherResult struct {
	EventID string
}

func CreateTeacherCommandHandler(ctx context.Context, command CreateTeacherCommand, saver eventstore.Saver) (CreateTeacherResult, error) {
	model, err := newCreateTeacherContext(command)
	if err != nil {
		return CreateTeacherResult{}, err
	}

	event := NewTeacherCreatedEvent(model.id, model.givenName, model.chosenName, model.familyName, time.Now(), metadataWithQuery(command.Metadata, model.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateTeacherResult{}, err
	}
	return CreateTeacherResult{EventID: model.id}, nil
}

type createTeacherContext struct {
	id         string
	givenName  string
	chosenName string
	familyName string
	query      eventstore.Query
}

func newCreateTeacherContext(command CreateTeacherCommand) (*createTeacherContext, error) {
	id := uuidv7.NewString()
	return &createTeacherContext{
		id:         id,
		givenName:  command.GivenName,
		chosenName: command.ChosenName,
		familyName: command.FamilyName,
		query:      streamQuery(id),
	}, nil
}
