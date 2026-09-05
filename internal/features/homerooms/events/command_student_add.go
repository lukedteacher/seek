package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	studentEvents "seek/internal/features/students/events"
)

type AddStudentToHomeroomCommand struct {
	HomeroomID string
	StudentID  string
	Metadata   CommandMetadata
}

type AddStudentToHomeroomResult struct {
	EventID string
	Skipped bool
}

func AddStudentToHomeroomCommandHandler(
	ctx context.Context,
	cmd AddStudentToHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddStudentToHomeroomResult,
	error,
) {
	model, err := loadAddStudentToHomeroomContext(ctx, retriever, cmd.HomeroomID, cmd.StudentID)
	if err != nil {
		return nil, err
	}
	if err := model.isHomeroomActive(); err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}
	skip := model.student.added
	if skip {
		return &AddStudentToHomeroomResult{Skipped: skip}, nil
	}

	event := NewStudentAddedToHomeroomEvent(
		cmd.HomeroomID,
		cmd.StudentID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
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
	return &AddStudentToHomeroomResult{EventID: event.EventID, Skipped: false}, nil
}

type studentState struct {
	created  bool
	archived bool
	deleted  bool
	added    bool
}

type addStudentToHomeroomContext struct {
	homeroom homeroomState
	student  studentState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadAddStudentToHomeroomContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	homeroomID,
	studentID string,
) (
	*addStudentToHomeroomContext,
	error,
) {
	homeroomStudentQuery := homeroomStudentStreamQuery(homeroomID, studentID)
	studentQuery := studentEvents.StreamQuery(studentID)
	query := combineQueries(homeroomStudentQuery, studentQuery)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &addStudentToHomeroomContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *addStudentToHomeroomContext) isHomeroomActive() error {
	if !m.homeroom.created || m.homeroom.archived || m.homeroom.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func (m *addStudentToHomeroomContext) isStudentActive() error {
	if !m.student.created || m.student.archived || m.student.deleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *addStudentToHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.homeroom.created = true
	case EventHomeroomArchived:
		m.homeroom.archived = true
	case EventHomeroomDeleted:
		m.homeroom.deleted = true
	case studentEvents.EventStudentCreated:
		m.student.created = true
	case studentEvents.EventStudentArchived:
		m.student.archived = true
	case studentEvents.EventStudentDeleted:
		m.student.deleted = true
	case EventStudentAddedToHomeroom:
		m.student.added = true
	case EventStudentRemovedFromHomeroom:
		m.student.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
