package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"seek/pkg/uuidv7"

	"seek/internal/eventstore"
)

type CreateEducatorCommand struct {
	GivenName  string
	ChosenName string
	FamilyName string
	Email      string
	Role       string
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
	eventID := uuidv7.NewString()
	model, err := newCreateEducatorContext(command, eventID)
	if err != nil {
		return CreateEducatorResult{}, fmt.Errorf("building educator context: %w", err)
	}

	// build event data struct directly
	eventData := EducatorCreatedEvent{
		EventID: eventID,
		EducatorState: EducatorState{
			GivenName:  model.givenName,
			ChosenName: model.chosenName,
			FamilyName: model.familyName,
			Email:      model.email,
			Username:   model.username,
			Role:       model.role,
		},
		CreatedAt: time.Now(),
		Scope:     educatorScope(model.id),
	}

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorCreated,
		Data:      eventstore.MustData(eventData),
		Metadata:  metadataWithQuery(command.Metadata, model.query),
	}

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
	username   string
	role       string
	query      eventstore.Query
}

// minimal logic here, since the educator is its own event root
// TODO consider creating an educator model and using that for validation
func newCreateEducatorContext(command CreateEducatorCommand, eventID string) (*createEducatorContext, error) {
	username := deriveUsername(command.Email)
	return &createEducatorContext{
		id:         eventID,
		givenName:  command.GivenName,
		chosenName: command.ChosenName,
		familyName: command.FamilyName,
		email:      command.Email,
		username:   username,
		role:       command.Role,
		query:      streamQuery(eventID),
	}, nil
}

func deriveUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	username := strings.ReplaceAll(localPart, ".", "")
	return strings.ToLower(username)
}
