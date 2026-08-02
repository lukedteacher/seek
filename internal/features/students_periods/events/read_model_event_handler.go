package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	student "seek/internal/features/students/events"
)

const PeriodStudentReadModelEventHandlerName = "period_student_read_model_event_handler"

type StudentAddedToPeriodProjection struct {
	Position  eventstore.Position
	PeriodID  string
	StudentID string
	AddedAt   time.Time
}

type StudentRemovedFromPeriodProjection struct {
	Position  eventstore.Position
	PeriodID  string
	StudentID string
	RemovedAt time.Time
}

type PeriodStudentReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel PeriodStudentReadModelWriter
	publisher eventstore.Publisher
}

func NewPeriodStudentReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel PeriodStudentReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*PeriodStudentReadModelEventHandler,
	error,
) {
	handler := &PeriodStudentReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            PeriodStudentReadModelEventHandlerName,
		Query:           ScheduleReadModelEventHandlerQuery(),
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

func (h *PeriodStudentReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *PeriodStudentReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func ScheduleReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		StudentAddedToPeriod,
		StudentRemovedFromPeriod,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodStudentReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	var periodID, studentID string

	switch resolved.Event.EventType {
	case StudentAddedToPeriod:
		var event StudentAddedToPeriodEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		periodID = event.PeriodID
		studentID = event.StudentID
		if err := h.readModel.AddStudentToPeriod(ctx, StudentAddedToPeriodProjection{
			Position:  resolved.Position,
			PeriodID:  event.PeriodID,
			StudentID: event.StudentID,
			AddedAt:   event.AddedAt,
		}); err != nil {
			return err
		}
	case StudentRemovedFromPeriod:
		var event StudentRemovedFromPeriodEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		periodID = event.PeriodID
		studentID = event.StudentID
		if err := h.readModel.RemoveStudentFromPeriod(ctx, StudentRemovedFromPeriodProjection{
			Position:  resolved.Position,
			PeriodID:  event.PeriodID,
			StudentID: event.StudentID,
			RemovedAt: event.RemovedAt,
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period student read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	_ = h.publisher.Publish(ctx, period.Channel(periodID), "period student read model update")
	_ = h.publisher.Publish(ctx, student.Channel(studentID), "period student read model update")
	return nil
}
