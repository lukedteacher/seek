package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/student"
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

	skipped := false
	if len(model.studentIDs) > 0 {
		for _, studentID := range model.studentIDs {
			if studentID == command.StudentID {
				println("skipped is true")
				skipped = true
			}
		}
	}
	if skipped {
		return &PeriodStudentAddResult{Skipped: skipped}, nil
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
	studentIDs     []string
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
	println("events length: ", len(events))
	if err != nil {
		return nil, err
	}

	model := &periodStudentAddContext{position: eventstore.NoEventPosition, events: events, query: query}
	println("criteria length: ", len(model.query.Criteria))
	for _, event := range events {
		println("event type: ", event.Event.EventType)
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
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case PeriodCreated:
		c.periodCreated = true
		c.periodDeleted = false
	case PeriodDeleted:
		c.periodDeleted = true
	case StudentCreated:
		c.studentCreated = true
		c.studentDeleted = false
	case StudentDeleted:
		c.studentDeleted = true
	case PeriodStudentAdded:
		// TODO could this be easier?
		studentID := data[student.StudentIDField].(string)
		c.studentIDs = append(c.studentIDs, studentID)
		println("handle, case psa: ", len(c.studentIDs))
	case PeriodStudentRemoved:
		studentIDToRemove := data[student.StudentIDField].(string)
		n := 0
		for index, studentID := range c.studentIDs {
			if c.studentIDs[index] == studentIDToRemove {
				c.studentIDs[n] = studentID
				n++
			}
		}
		c.studentIDs = c.studentIDs[:n]
		println("handle, case psr: ", len(c.studentIDs))
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
