package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type AddIEPToStudentCommand struct {
	IEP      IEPState
	Metadata CommandMetadata
}

type AddStudentIEPToStudentResult struct {
	EventID string
	Skipped bool
}

func AddIEPToStudentCommandHandler(
	ctx context.Context,
	command AddIEPToStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddStudentIEPToStudentResult,
	error,
) {
	model, err := loadAddStudentIEPToStudentContext(
		ctx,
		retriever,
		command.IEP.StudentID,
	)
	if err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}
	eventID := uuidv7.NewString()
	event := NewStudentIEPAddedToStudentEvent(
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
	return &AddStudentIEPToStudentResult{EventID: eventID, Skipped: false}, nil
}

type addStudentIEPToStudentContext struct {
	studentCreated  bool
	studentArchived bool
	studentDeleted  bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadAddStudentIEPToStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (
	*addStudentIEPToStudentContext,
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
	model := &addStudentIEPToStudentContext{
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

func (m *addStudentIEPToStudentContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *addStudentIEPToStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case events.EventStudentCreated:
		m.studentCreated = true
		m.studentArchived = false
		m.studentDeleted = false
	case events.EventStudentArchived:
		m.studentArchived = true
	case events.EventStudentDeleted:
		m.studentDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
