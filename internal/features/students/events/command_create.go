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
	Grade       int
	Homeroom    string
	CaseManager string
	Metadata    CommandMetadata
}

type CreateStudentResult struct {
	EventID string
}

func CreateStudentCommandHandler(
	ctx context.Context,
	command CreateStudentCommand,
	saver eventstore.Saver,
) (
	CreateStudentResult,
	error,
) {
	eventID := uuidv7.NewString()
	model, err := newCreateStudentContext(command, eventID)
	if err != nil {
		return CreateStudentResult{}, err
	}
	event := NewStudentCreatedEvent(
		eventID,
		model.givenName,
		model.chosenName,
		model.familyName,
		model.grade,
		model.homeroom,
		model.caseManager,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateStudentResult{}, err
	}
	return CreateStudentResult{EventID: eventID}, nil
}

type createStudentContext struct {
	givenName   string
	chosenName  string
	familyName  string
	grade       int
	homeroom    string
	caseManager string
	query       eventstore.Query
}

func newCreateStudentContext(command CreateStudentCommand, eventID string) (*createStudentContext, error) {
	return &createStudentContext{
		givenName:   command.GivenName,
		chosenName:  command.ChosenName,
		familyName:  command.FamilyName,
		grade:       command.Grade,
		homeroom:    command.Homeroom,
		caseManager: command.CaseManager,
		query:       StreamQuery(eventID),
	}, nil
}
