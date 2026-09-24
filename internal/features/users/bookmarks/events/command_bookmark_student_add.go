package events

import (
	"context"
	"seek/internal/auth"
	"seek/internal/eventstore"
	studentEvents "seek/internal/features/students/events"
)

type AddStudentBookmarkCommand struct {
	UserID    string
	StudentID string
	Metadata  CommandMetadata
}

type AddStudentBookmarkResult struct {
	EventID string
	Skipped bool
}

func AddStudentBookmarkCommandHandler(
	ctx context.Context,
	cmd AddStudentBookmarkCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	AddStudentBookmarkResult,
	error,
) {
	model, err := loadAddStudentBookmarkContext(
		ctx,
		retriever,
		cmd.UserID,
		cmd.StudentID,
	)
	if err != nil {
		return AddStudentBookmarkResult{}, err
	}
	if !model.user.registered || model.user.deleted {
		return AddStudentBookmarkResult{}, eventstore.ErrUserNotActive
	}
	if !model.student.created || model.student.archived || model.student.deleted {
		return AddStudentBookmarkResult{}, eventstore.ErrStudentNotActive
	}
	if model.added {
		return AddStudentBookmarkResult{
			Skipped: true,
		}, err
	}
	event := NewStudentBookmarkAddedEvent(
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
		return AddStudentBookmarkResult{}, err
	}
	return AddStudentBookmarkResult{
		EventID: event.EventID,
		Skipped: false,
	}, nil
}

type addStudentBookmarkContext struct {
	user     UserState
	student  StudentState
	added    bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadAddStudentBookmarkContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	userID,
	studentID string,
) (
	addStudentBookmarkContext,
	error,
) {
	query := studentBookmarkStreamQuery(userID, studentID)
	events, err := retriever.GetEvents(
		ctx,
		eventstore.NoEventPosition,
		100,
		eventstore.Forward,
		query,
	)
	if err != nil {
		return addStudentBookmarkContext{}, err
	}
	model := addStudentBookmarkContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *addStudentBookmarkContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case auth.UserRegistered:
		m.user.registered = true
	case auth.AccountDeleted:
		m.user.deleted = true
	case studentEvents.EventStudentCreated:
		m.student.created = true
	case studentEvents.EventStudentArchived:
		m.student.archived = true
	case studentEvents.EventStudentDeleted:
		m.student.deleted = true
	case EventStudentBookmarkAdded:
		m.added = true
	case EventStudentBookmarkRemoved:
		m.added = false
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
