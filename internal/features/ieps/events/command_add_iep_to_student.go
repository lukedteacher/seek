package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	studentEvents "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type AddIEPToStudentCommand struct {
	IEPState
	Metadata CommandMetadata
}

type AddStudentIEPToStudentResult struct {
	EventID string
	Skipped bool
}

func AddIEPToStudentCommandHandler(
	ctx context.Context,
	cmd AddIEPToStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	AddStudentIEPToStudentResult,
	error,
) {
	model, err := loadAddStudentIEPToStudentContext(
		ctx,
		retriever,
		cmd.IEPState.StudentID,
	)
	if err != nil {
		return AddStudentIEPToStudentResult{}, err
	}
	if err := model.isStudentActive(); err != nil {
		return AddStudentIEPToStudentResult{}, err
	}
	if model.student.hasActiveIEP {
		return AddStudentIEPToStudentResult{}, eventstore.ErrIEPStudentHasActiveIEP
	}
	eventID := uuidv7.NewString()
	cmd.IEPState.ID = eventID
	event := NewIEPAddedToStudentEvent(
		cmd.IEPState,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		nil,
		model.query,
	); err != nil {
		return AddStudentIEPToStudentResult{}, err
	}
	return AddStudentIEPToStudentResult{EventID: eventID, Skipped: false}, nil
}

type addStudentIEPToStudentContext struct {
	student  StudentState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadAddStudentIEPToStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (
	*addStudentIEPToStudentContext,
	error,
) {
	query := studentStreamQuery(studentID)
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
	if !m.student.isCreated || m.student.isArchived || m.student.isDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *addStudentIEPToStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case studentEvents.EventStudentCreated:
		m.student.isCreated = true
	case studentEvents.EventStudentArchived:
		m.student.isArchived = true
	case studentEvents.EventStudentDeleted:
		m.student.isDeleted = true
	case EventIEPAddedToStudent:
		m.student.hasActiveIEP = true
	case EventIEPArchived:
		m.student.hasActiveIEP = false
	case EventIEPDeleted:
		m.student.hasActiveIEP = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
