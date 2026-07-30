package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type ArchiveStudentCommand struct {
	StudentID string
	Metadata  CommandMetadata
}

type ArchiveStudentResult struct {
	EventID string
}

func ArchiveStudentCommandHandler(
	ctx context.Context,
	command ArchiveStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	ArchiveStudentResult,
	error,
) {
	model, err := loadArchiveStudentContext(ctx, retriever, command.StudentID)
	if err != nil {
		return ArchiveStudentResult{}, err
	}
	if err := model.isActive(); err != nil {
		return ArchiveStudentResult{}, err
	}

	eventID := uuidv7.NewString()
	event := NewStudentArchivedEvent(
		eventID,
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
		return ArchiveStudentResult{}, err
	}
	return ArchiveStudentResult{EventID: eventID}, nil
}

type archiveStudentContext struct {
	created  bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadArchiveStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (
	*archiveStudentContext,
	error,
) {
	query := StreamQuery(studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &archiveStudentContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *archiveStudentContext) isActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrNotActive
	}
	return nil
}

func (m *archiveStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case StudentCreated:
		m.created = true
		m.archived = false
		m.deleted = false
	case StudentArchived:
		m.archived = true
	case StudentDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
