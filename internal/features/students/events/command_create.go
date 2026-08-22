package events

import (
	"context"
	"strings"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type CreateStudentCommand struct {
	StudentState
	Metadata CommandMetadata
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
		model.StudentState,
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
		return CreateStudentResult{}, err
	}
	return CreateStudentResult{EventID: model.ID}, nil
}

type createStudentContext struct {
	StudentState
	query eventstore.Query
}

func newCreateStudentContext(command CreateStudentCommand) (*createStudentContext, error) {
	id := uuidv7.NewString()
	username := deriveUsername(command.Email)
	context := &createStudentContext{
		StudentState: command.StudentState,
		query:        StreamQuery(id),
	}
	context.ID = id
	context.Username = username
	return context, nil
}

func deriveUsername(email string) string {
	localPart := strings.Split(email, "@")[0]
	username := strings.ReplaceAll(localPart, ".", "")
	return strings.ToLower(username)
}
