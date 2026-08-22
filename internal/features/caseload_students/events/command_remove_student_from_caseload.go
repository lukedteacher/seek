package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	ee "seek/internal/features/educators/events"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type RemoveStudentFromCaseloadCommand struct {
	EducatorID string
	StudentID  string
	Metadata   CommandMetadata
}

type RemoveStudentFromCaseloadResult struct {
	EventID string
	Skipped bool
}

func RemoveStudentFromCaseloadCommandHandler(
	ctx context.Context,
	command RemoveStudentFromCaseloadCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*RemoveStudentFromCaseloadResult,
	error,
) {
	model, err := loadRemoveStudentFromCaseloadContext(ctx, retriever, command.EducatorID, command.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isEducatorActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}

	skip := !model.added
	if skip {
		return &RemoveStudentFromCaseloadResult{Skipped: skip}, nil
	}
	eventID := uuidv7.NewString()
	event := NewStudentRemovedFromCaseloadEvent(
		eventID,
		command.EducatorID,
		command.StudentID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return nil, err
	}
	return &RemoveStudentFromCaseloadResult{EventID: eventID, Skipped: false}, nil
}

type removeStudentFromEducatorContext struct {
	educatorCreated  bool
	educatorArchived bool
	educatorDeleted  bool
	studentCreated   bool
	studentArchived  bool
	studentDeleted   bool
	added            bool
	position         eventstore.Position
	events           []eventstore.ResolvedEvent
	query            eventstore.Query
}

func loadRemoveStudentFromCaseloadContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	educatorID string,
	studentID string,
) (
	*removeStudentFromEducatorContext,
	error,
) {
	query := educatorStudentStreamQuery(educatorID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &removeStudentFromEducatorContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *removeStudentFromEducatorContext) isEducatorActive() error {
	if !m.educatorCreated || m.educatorArchived || m.educatorDeleted {
		return eventstore.ErrEducatorNotFound
	}
	return nil
}

func (m *removeStudentFromEducatorContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrStudentNotFound
	}
	return nil
}

func (m *removeStudentFromEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case ee.EventEducatorCreated:
		m.educatorCreated = true
	case ee.EventEducatorArchived:
		m.educatorArchived = true
	case ee.EventEducatorDeleted:
		m.educatorDeleted = true
	case se.EventStudentCreated:
		m.studentCreated = true
	case se.EventStudentArchived:
		m.studentArchived = true
	case se.EventStudentDeleted:
		m.studentDeleted = true
	case StudentAddedToCaseload:
		m.added = true
	case StudentRemovedFromCaseload:
		m.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
