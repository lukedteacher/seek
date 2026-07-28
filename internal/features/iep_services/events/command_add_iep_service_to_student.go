package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type AddIEPServiceToStudentCommand struct {
	StudentID       string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
	Metadata        CommandMetadata
}

type AddIEPServiceToStudentResult struct {
	EventID string
	Skipped bool
}

func AddIEPServiceToStudentCommandHandler(
	ctx context.Context,
	command AddIEPServiceToStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddIEPServiceToStudentResult,
	error,
) {
	model, err := loadAddIEPServiceToStudentContext(
		ctx,
		retriever,
		command.StudentID,
	)
	if err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}
	eventID := uuidv7.NewString()
	event := NewIEPServiceAddedToStudentEvent(
		eventID,
		command,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		nil,
		model.query,
	); err != nil {
		return nil, err
	}
	return &AddIEPServiceToStudentResult{EventID: eventID, Skipped: false}, nil
}

type addIEPServiceToStudentContext struct {
	studentCreated  bool
	studentArchived bool
	studentDeleted  bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadAddIEPServiceToStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (
	*addIEPServiceToStudentContext,
	error,
) {
	query := events.StreamQuery(studentID)
	events, err := retriever.GetEvents(
		ctx,
		eventstore.NoEventPosition,
		100,
		eventstore.Forward,
		query,
	)
	if err != nil {
		return nil, err
	}
	model := &addIEPServiceToStudentContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	// creates a model of the relevant context from past events
	for i := range events {
		model.handle(events[i])
	}
	return model, nil
}

func (m *addIEPServiceToStudentContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *addIEPServiceToStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case events.StudentCreated:
		m.studentCreated = true
		m.studentArchived = false
		m.studentDeleted = false
	case events.StudentArchived:
		m.studentArchived = true
	case events.StudentDeleted:
		m.studentDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
