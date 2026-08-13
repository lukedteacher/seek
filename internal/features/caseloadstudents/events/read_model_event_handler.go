package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	educator "seek/internal/features/educators/events"
	student "seek/internal/features/students/events"
)

const CaseManagerStudentReadModelEventHandlerName = "caseload_student_read_model_event_handler"

type StudentAddedToCaseloadProjection struct {
	Position   eventstore.Position
	EducatorID string
	StudentID  string
	AddedAt    time.Time
}

type StudentRemovedFromCaseloadProjection struct {
	Position   eventstore.Position
	EducatorID string
	StudentID  string
	RemovedAt  time.Time
}

type CaseManagerStudentReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel CaseManagerStudentReadModelWriter
	publisher eventstore.Publisher
}

func NewCaseloadStudentReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel CaseManagerStudentReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*CaseManagerStudentReadModelEventHandler,
	error,
) {
	handler := &CaseManagerStudentReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            CaseManagerStudentReadModelEventHandlerName,
		Query:           CaseloadStudentReadModelEventHandlerQuery(),
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

func (h *CaseManagerStudentReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *CaseManagerStudentReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func CaseloadStudentReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		StudentAddedToCaseload,
		StudentRemovedFromCaseload,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *CaseManagerStudentReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	var educatorID, studentID string
	switch resolved.Event.EventType {
	case StudentAddedToCaseload:
		var event StudentAddedToCaseloadEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		educatorID = event.EducatorID
		studentID = event.StudentID
		if err := h.readModel.AddStudentToCaseload(ctx, StudentAddedToCaseloadProjection{
			Position:   resolved.Position,
			EducatorID: event.EducatorID,
			StudentID:  event.StudentID,
			AddedAt:    event.AddedAt,
		}); err != nil {
			return err
		}
	case StudentRemovedFromCaseload:
		var event StudentRemovedFromCaseloadEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		educatorID = event.EducatorID
		studentID = event.StudentID
		if err := h.readModel.RemoveStudentFromCaseload(ctx, StudentRemovedFromCaseloadProjection{
			Position:   resolved.Position,
			EducatorID: event.EducatorID,
			StudentID:  event.StudentID,
			RemovedAt:  event.RemovedAt,
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period educator read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	_ = h.publisher.Publish(ctx, educator.Channel(educatorID), "case manager student educator read model update")
	_ = h.publisher.Publish(ctx, student.Channel(studentID), "case manager student student read model update")
	return nil
}
