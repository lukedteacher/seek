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
	EducatorState
	Metadata CommandMetadata
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
		EventID:       eventID,
		EducatorState: model.EducatorState,
		Scope:         educatorScope(eventID),
	}
	eventData.CreatedAt = time.Now()

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorCreated,
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

	return CreateEducatorResult{EventID: eventID}, nil
}

type createEducatorContext struct {
	EducatorState
	query eventstore.Query
}

// minimal logic here, since the educator is its own event root
// TODO consider creating an educator model and using that for validation
func newCreateEducatorContext(command CreateEducatorCommand, eventID string) (*createEducatorContext, error) {
	username := deriveUsername(command.Email)
	educator := createEducatorContext{
		EducatorState: command.EducatorState,
		query:         streamQuery(eventID),
	}
	educator.ID = eventID
	educator.Username = username
	return &educator, nil
}

func deriveUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	username := strings.ReplaceAll(localPart, ".", "")
	return strings.ToLower(username)
}
