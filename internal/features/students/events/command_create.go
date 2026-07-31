package events

import (
	"context"
	"strings"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type CreateStudentCommand struct {
	MARSSID     string
	GivenName   string
	ChosenName  string
	FamilyName  string
	Email       string
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
	model, err := newCreateStudentContext(command)
	if err != nil {
		return CreateStudentResult{}, err
	}
	event := NewStudentCreatedEvent(
		model.id,
		model.marssID,
		model.givenName,
		model.chosenName,
		model.familyName,
		model.email,
		model.username,
		model.grade,
		model.homeroom,
		model.caseManager,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, model.query); err != nil {
		return CreateStudentResult{}, err
	}
	return CreateStudentResult{EventID: model.id}, nil
}

type createStudentContext struct {
	id          string
	marssID     string
	givenName   string
	chosenName  string
	familyName  string
	email       string
	username    string
	grade       int
	homeroom    string
	caseManager string
	query       eventstore.Query
}

func newCreateStudentContext(command CreateStudentCommand) (*createStudentContext, error) {
	id := uuidv7.NewString()
	username := deriveUsername(command.Email)
	return &createStudentContext{
		id:          id,
		marssID:     command.MARSSID,
		givenName:   command.GivenName,
		chosenName:  command.ChosenName,
		familyName:  command.FamilyName,
		email:       command.Email,
		username:    username,
		grade:       command.Grade,
		homeroom:    command.Homeroom,
		caseManager: command.CaseManager,
		query:       StreamQuery(id),
	}, nil
}

func deriveUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	username := strings.ReplaceAll(localPart, ".", "")
	return strings.ToLower(username)
}
