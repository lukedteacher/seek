package events

import (
	"context"
	"seek/internal/auth"
	"seek/internal/eventstore"
	studentEvents "seek/internal/features/students/events"
)

type RemoveStudentBookmarkCommand struct {
	UserID    string
	StudentID string
	Metadata  CommandMetadata
}

type RemoveStudentBookmarkResult struct {
	EventID string
	Skipped bool
}

func RemoveStudentBookmarkCommandHandler(
	ctx context.Context,
	cmd RemoveStudentBookmarkCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	RemoveStudentBookmarkResult,
	error,
) {
	model, err := loadRemoveStudentBookmarkContext(
		ctx,
		retriever,
		cmd.UserID,
		cmd.StudentID,
	)
	if err != nil {
		return RemoveStudentBookmarkResult{}, err
	}
	if !model.user.registered || model.user.deleted {
		return RemoveStudentBookmarkResult{}, eventstore.ErrUserNotActive
	}
	if !model.student.created || model.student.archived || model.student.deleted {
		return RemoveStudentBookmarkResult{}, eventstore.ErrStudentNotActive
	}
	if !model.added {
		return RemoveStudentBookmarkResult{
			Skipped: true,
		}, err
	}
	event := NewStudentBookmarkRemovedEvent(
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
		return RemoveStudentBookmarkResult{}, err
	}
	return RemoveStudentBookmarkResult{
		EventID: event.EventID,
		Skipped: false,
	}, nil
}

type removeStudentBookmarkContext struct {
	user     UserState
	student  StudentState
	added    bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadRemoveStudentBookmarkContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	userID,
	studentID string,
) (
	removeStudentBookmarkContext,
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
		return removeStudentBookmarkContext{}, err
	}
	model := removeStudentBookmarkContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	for _, event := range events {
		model.handle(event)
	}
	println("q", len(query.Criteria), "e", len(events), "u", model.user.registered)
	return model, nil
}

func (m *removeStudentBookmarkContext) handle(resolved eventstore.ResolvedEvent) {
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
