package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type RemoveStudentFromPeriodCommand struct {
	PeriodID  string
	StudentID string
	Metadata  CommandMetadata
}

type RemoveStudentFromPeriodResult struct {
	EventID string
	Skipped bool
}

func RemoveStudentFromPeriodCommandHandler(
	ctx context.Context,
	command RemoveStudentFromPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*RemoveStudentFromPeriodResult,
	error,
) {
	model, err := loadRemoveStudentFromPeriodContext(ctx, retriever, command.PeriodID, command.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}

	skip := !model.added
	if skip {
		return &RemoveStudentFromPeriodResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewStudentRemovedFromPeriodEvent(eventID, command.PeriodID, command.StudentID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &RemoveStudentFromPeriodResult{EventID: eventID, Skipped: false}, nil
}

type removeStudentFromPeriodContext struct {
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

func loadRemoveStudentFromPeriodContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	studentID string,
) (
	*removeStudentFromPeriodContext,
	error,
) {
	query := studentPeriodStreamQuery(periodID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &removeStudentFromPeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *removeStudentFromPeriodContext) isPeriodActive() error {
	if !m.periodCreated || m.periodArchived || m.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *removeStudentFromPeriodContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrStudentNotFound
	}
	return nil
}

func (m *removeStudentFromPeriodContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		m.periodCreated = true
	case pe.PeriodArchived:
		m.periodArchived = true
	case pe.PeriodDeleted:
		m.periodDeleted = true
	case se.StudentCreated:
		m.studentCreated = true
	case se.StudentArchived:
		m.studentArchived = true
	case se.StudentDeleted:
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
