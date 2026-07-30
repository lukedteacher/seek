package events

import (
	"context"
	"strings"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
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
	eventID := uuidv7.NewString()
	model, err := newCreateStudentContext(command, eventID)
	if err != nil {
		return CreateStudentResult{}, err
	}
	event := NewStudentCreatedEvent(
		eventID,
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
	return CreateStudentResult{EventID: eventID}, nil
}

type createStudentContext struct {
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

func newCreateStudentContext(command CreateStudentCommand, eventID string) (*createStudentContext, error) {
	username := deriveUsername(command.Email)
	return &createStudentContext{
		marssID:     command.MARSSID,
		givenName:   command.GivenName,
		chosenName:  command.ChosenName,
		familyName:  command.FamilyName,
		email:       command.Email,
		username:    username,
		grade:       command.Grade,
		homeroom:    command.Homeroom,
		caseManager: command.CaseManager,
		query:       StreamQuery(eventID),
	}, nil
}

func deriveUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	username := strings.ReplaceAll(localPart, ".", "")
	return strings.ToLower(username)
}
