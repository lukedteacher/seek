package events

import (
	"context"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreateStudentCommand struct {
	GivenName   string
	ChosenName  string
	FamilyName  string
	Grade       int64
	Homeroom    string
	CaseManager string
	Metadata    CommandMetadata
}

type CreateStudentResult struct {
	Id string
}

func CreateStudentCommandHandler(ctx context.Context, command CreateStudentCommand, saver eventstore.Saver) (CreateStudentResult, error) {
	model, err := newCreateStudentContext(command)
	if err != nil {
		return CreateStudentResult{}, err
	}

	event := NewStudentCreatedEvent(model.id, model.firstName, model.chosenName, model.lastName, model.grade, model.homeroom, model.caseManager, time.Now(), metadataWithQuery(command.Metadata, model.query))
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateStudentResult{}, err
	}
	return CreateStudentResult{Id: model.id}, nil
}

type createStudentContext struct {
	id          string
	firstName   string
	chosenName  string
	lastName    string
	grade       int64
	homeroom    string
	caseManager string
	query       eventstore.Query
}

func newCreateStudentContext(command CreateStudentCommand) (*createStudentContext, error) {
	id := uuidv7.NewString()
	return &createStudentContext{
		id:          id,
		firstName:   command.GivenName,
		chosenName:  command.ChosenName,
		lastName:    command.FamilyName,
		grade:       command.Grade,
		homeroom:    command.Homeroom,
		caseManager: command.CaseManager,
		query:       StreamQuery(id),
	}, nil
}
