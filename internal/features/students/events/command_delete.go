package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type DeleteStudentCommand struct {
	StudentID string
	Metadata  CommandMetadata
}

type DeleteStudentResult struct {
	EventID string
}

func DeleteStudentCommandHandler(ctx context.Context, command DeleteStudentCommand, saver eventstore.Saver, retriever eventstore.Retriever) (DeleteStudentResult, error) {
	model, err := loadDeleteStudentContext(ctx, retriever, command.StudentID)
	if err != nil {
		return DeleteStudentResult{}, err
	}
	if !model.isActive() {
		return DeleteStudentResult{}, eventstore.ErrStudentNotActive
	}

	eventID := uuidv7.NewString()
	event := NewStudentDeletedEvent(eventID, command.StudentID, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteStudentResult{}, err
	}
	return DeleteStudentResult{EventID: eventID}, nil
}

type deleteStudentContext struct {
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteStudentContext(ctx context.Context, retriever eventstore.Retriever, studentID string) (*deleteStudentContext, error) {
	query := StreamQuery(studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteStudentContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteStudentContext) isActive() bool {
	if !m.exists || m.deleted {
		return false
	}
	return true
}

func (m *deleteStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case StudentCreated:
		m.exists = true
		m.deleted = false
	case StudentDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
