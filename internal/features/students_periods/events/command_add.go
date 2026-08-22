package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type AddStudentToPeriodCommand struct {
	PeriodID  string
	StudentID string
	Metadata  CommandMetadata
}

type AddStudentToPeriodResult struct {
	EventID string
	Skipped bool
}

func AddStudentToPeriodCommandHandler(
	ctx context.Context,
	command AddStudentToPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddStudentToPeriodResult,
	error,
) {
	model, err := loadAddStudentToPeriodContext(ctx, retriever, command.PeriodID, command.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}

	skip := model.added
	if skip {
		return &AddStudentToPeriodResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewStudentAddedToPeriodEvent(
		eventID,
		command.PeriodID,
		command.StudentID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		model.events,
		model.query,
	); err != nil {
		return nil, err
	}
	return &AddStudentToPeriodResult{EventID: eventID, Skipped: false}, nil
}

type addStudentToPeriodContext struct {
	periodCreated   bool
	periodArchived  bool
	periodDeleted   bool
	studentCreated  bool
	studentArchived bool
	studentDeleted  bool
	added           bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadAddStudentToPeriodContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	studentID string,
) (
	*addStudentToPeriodContext,
	error,
) {
	query := studentPeriodStreamQuery(periodID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &addStudentToPeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *addStudentToPeriodContext) isPeriodActive() error {
	if !m.periodCreated || m.periodArchived || m.periodDeleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func (m *addStudentToPeriodContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *addStudentToPeriodContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case pe.EventPeriodCreated:
		m.periodCreated = true
	case pe.EventPeriodArchived:
		m.periodArchived = true
	case pe.EventPeriodDeleted:
		m.periodDeleted = true
	case se.EventStudentCreated:
		m.studentCreated = true
	case se.EventStudentArchived:
		m.studentArchived = true
	case se.EventStudentDeleted:
		m.studentDeleted = true
	case StudentAddedToPeriod:
		m.added = true
	case StudentRemovedFromPeriod:
		m.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
