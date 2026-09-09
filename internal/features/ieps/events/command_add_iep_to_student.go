package events

import (
	"context"

	"seek/internal/eventstore"
	"seek/internal/features/ieps/models"
	studentEvents "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type AddIEPToStudentCommand struct {
	IEP      models.IEP
	Metadata CommandMetadata
}

type AddIEPToStudentResult struct {
	EventID string
	Skipped bool
}

func AddIEPToStudentCommandHandler(
	ctx context.Context,
	cmd AddIEPToStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddIEPToStudentResult,
	error,
) {
	model, err := loadaddIEPToStudentContext(
		ctx,
		retriever,
		cmd.IEP.StudentID,
	)
	if err != nil {
		return &AddIEPToStudentResult{}, err
	}
	if !model.student.isActive() {
		return &AddIEPToStudentResult{}, eventstore.ErrStudentNotActive
	}
	if model.student.hasIEP() {
		return &AddIEPToStudentResult{}, eventstore.ErrIEPStudentHasActiveIEP
	}
	eventID := uuidv7.NewString()
	cmd.IEP.ID = eventID
	event := NewIEPAddedToStudentEvent(
		cmd,
		model.query,
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		nil,
		model.query,
	); err != nil {
		return &AddIEPToStudentResult{}, err
	}
	return &AddIEPToStudentResult{EventID: eventID, Skipped: false}, nil
}

type addIEPToStudentContext struct {
	student  StudentState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadaddIEPToStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (
	*addIEPToStudentContext,
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
	model := &addIEPToStudentContext{
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

func (m *addIEPToStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case studentEvents.EventStudentCreated:
		m.student.created = true
	case studentEvents.EventStudentArchived:
		m.student.archived = true
	case studentEvents.EventStudentDeleted:
		m.student.deleted = true
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
