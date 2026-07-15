package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
	"seek/internal/uuidv7"
)

type PeriodStudentAddCommand struct {
	EventID   string
	PeriodID  string
	StudentID string
	Metadata  CommandMetadata
}

type PeriodStudentAddResult struct {
	EventID string
	Skipped bool
}

func PeriodStudentAddCommandHandler(
	ctx context.Context,
	command PeriodStudentAddCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*PeriodStudentAddResult,
	error,
) {
	model, err := loadPeriodStudentAddContext(ctx, retriever, command.PeriodID, command.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}

	skip := model.studentAdded
	if skip {
		return &PeriodStudentAddResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewPeriodStudentAddedEvent(
		eventID,
		command.PeriodID,
		command.StudentID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &PeriodStudentAddResult{EventID: eventID, Skipped: false}, nil
}

type periodStudentAddContext struct {
	periodCreated  bool
	periodDeleted  bool
	studentCreated bool
	studentDeleted bool
	studentAdded   bool
	position       eventstore.Position
	events         []eventstore.ResolvedEvent
	query          eventstore.Query
}

func loadPeriodStudentAddContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	studentID string,
) (
	*periodStudentAddContext,
	error,
) {
	query := streamQuery(periodID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &periodStudentAddContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *periodStudentAddContext) isPeriodActive() error {
	if !c.periodCreated || c.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (c *periodStudentAddContext) isStudentActive() error {
	if !c.studentCreated || c.studentDeleted {
		return eventstore.ErrStudentNotFound
	}
	return nil
}

func (c *periodStudentAddContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		c.periodCreated = true
		c.periodDeleted = false
	case pe.PeriodDeleted:
		c.periodDeleted = true
	case se.StudentCreated:
		c.studentCreated = true
		c.studentDeleted = false
	case se.StudentDeleted:
		c.studentDeleted = true
	case PeriodStudentAdded:
		c.studentAdded = true
	case PeriodStudentRemoved:
		c.studentAdded = false
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
