package events

import (
	"context"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreateTeacherCommand struct {
	FirstName  string
	ChosenName string
	LastName   string
	Metadata   CommandMetadata
}

type CreateTeacherResult struct {
	Id string
}

func CreateTeacherCommandHandler(ctx context.Context, command CreateTeacherCommand, saver eventstore.Saver) (CreateTeacherResult, error) {
	model, err := newCreateTeacherContext(command)
	if err != nil {
		return CreateTeacherResult{}, err
	}

	event := NewTeacherCreatedEvent(model.id, model.firstName, model.chosenName, model.lastName, time.Now(), metadataWithQuery(command.Metadata, model.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateTeacherResult{}, err
	}
	return CreateTeacherResult{Id: model.id}, nil
}

type createTeacherContext struct {
	id         string
	firstName  string
	chosenName string
	lastName   string
	query      eventstore.Query
}

func newCreateTeacherContext(command CreateTeacherCommand) (*createTeacherContext, error) {
	id := uuidv7.NewString()
	return &createTeacherContext{
		id:         id,
		firstName:  command.FirstName,
		chosenName: command.ChosenName,
		lastName:   command.LastName,
		query:      streamQuery(id),
	}, nil
}
