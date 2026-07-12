package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type PeriodStudentRemoveCommand struct {
	EventID   string
	PeriodID  string
	StudentID string
	Metadata  CommandMetadata
}

type PeriodStudentRemoveResult struct {
	EventID string
	Skipped bool
}

func PeriodStudentRemoveCommandHandler(
	ctx context.Context,
	command PeriodStudentRemoveCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*PeriodStudentRemoveResult,
	error,
) {
	model, err := loadPeriodStudentRemoveContext(ctx, retriever, command.PeriodID, command.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isPeriodActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}

	skipped := true
	if len(model.studentIDs) > 0 {
		for _, studentID := range model.studentIDs {
			if studentID == command.StudentID {
				skipped = false
			}
		}
	}
	if skipped {
		return &PeriodStudentRemoveResult{Skipped: skipped}, nil
	}
	eventID := uuidv7.NewString()
	event := NewPeriodStudentRemovedEvent(eventID, command.PeriodID, command.StudentID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &PeriodStudentRemoveResult{EventID: eventID, Skipped: false}, nil
}

type periodStudentRemoveContext struct {
	periodCreated  bool
	periodDeleted  bool
	studentCreated bool
	studentDeleted bool
	studentIDs     []string
	position       eventstore.Position
	events         []eventstore.ResolvedEvent
	query          eventstore.Query
}

func loadPeriodStudentRemoveContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	periodID string,
	studentID string,
) (
	*periodStudentRemoveContext,
	error,
) {
	query := streamQuery(periodID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &periodStudentRemoveContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (c *periodStudentRemoveContext) isPeriodActive() error {
	if !c.periodCreated || c.periodDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (c *periodStudentRemoveContext) isStudentActive() error {
	if !c.studentCreated || c.studentDeleted {
		return eventstore.ErrStudentNotFound
	}
	return nil
}

func (c *periodStudentRemoveContext) handle(resolved eventstore.ResolvedEvent) {
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
		// TODO should there be more to this?
		studentID := data["student_id"].(string)
		c.studentIDs = append(c.studentIDs, studentID)
	case PeriodStudentRemoved:
		studentIDToRemove := data["studentID"].(string)
		n := 0
		for index, studentID := range c.studentIDs {
			if c.studentIDs[index] == studentIDToRemove {
				c.studentIDs[n] = studentID
				n++
			}
		}
		c.studentIDs = c.studentIDs[:n]
	}
	if resolved.Position.After(c.position) {
		c.position = resolved.Position
	}
}
