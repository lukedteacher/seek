package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	student "seek/internal/features/students/events"
)

const BookmarksReadModelEventHandlerName = "bookmarks_read_model_event_handler"

type BookmarkReadModelWriter interface {
	AddStudentBookmark(ctx context.Context, event StudentBookmarkAddedProjection) error
	RemoveStudentBookmark(ctx context.Context, event StudentBookmarkRemovedProjection) error
}

type StudentBookmarkAddedProjection struct {
	Position  eventstore.Position
	UserID    string
	StudentID string
	AddedAt   string
}

type StudentBookmarkRemovedProjection struct {
	Position  eventstore.Position
	UserID    string
	StudentID string
	RemovedAt time.Time
}

type BookmarkReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel BookmarkReadModelWriter
	publisher eventstore.Publisher
}

func NewReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel BookmarkReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*BookmarkReadModelEventHandler,
	error,
) {
	handler := &BookmarkReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            BookmarksReadModelEventHandlerName,
		Query:           BookmarkReadModelEventHandlerQuery(),
		Logger:          logger,
		MaxEventRetries: -1,
		HandleEvent:     handler.handle,
	})
	if err != nil {
		return nil, err
	}
	handler.global = global
	return handler, nil
}

func (h *BookmarkReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *BookmarkReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func BookmarkReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventStudentBookmarkAdded,
		EventStudentBookmarkRemoved,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *BookmarkReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	var userID, studentID string
	switch resolved.Event.EventType {
	case EventStudentBookmarkAdded:
		var event StudentBookmarkAddedEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		if err := h.readModel.AddStudentBookmark(ctx, StudentBookmarkAddedProjection{
			Position:  resolved.Position,
			UserID:    event.UserID,
			StudentID: event.StudentID,
			AddedAt:   event.AddedAt,
		}); err != nil {
			return err
		}
		userID = event.UserID
		studentID = event.StudentID
	case EventStudentBookmarkRemoved:
		var event StudentBookmarkRemovedEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		userID = event.UserID
		studentID = event.StudentID
		if err := h.readModel.RemoveStudentBookmark(ctx, StudentBookmarkRemovedProjection{
			Position:  resolved.Position,
			UserID:    event.UserID,
			StudentID: event.StudentID,
		}); err != nil {
			return err
		}
		userID = event.UserID
		studentID = event.StudentID
	default:
		return fmt.Errorf("unhandled period educator read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	_ = h.publisher.Publish(ctx, "users."+userID, "student bookmark user read model update")
	_ = h.publisher.Publish(ctx, student.Channel(studentID), "student bookmark student read model update")
	return nil
}
