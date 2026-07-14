package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	student "seek/internal/features/students/events"
)

const PeriodStudentReadModelEventHandlerName = "period_student_read_model_event_handler"

type PeriodStudentAddedProjection struct {
	Position  eventstore.Position
	PeriodID  string
	StudentID string
	AddedAt   time.Time
}

type PeriodStudentRemovedProjection struct {
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

func NewPeriodStudentReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel PeriodStudentReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*PeriodStudentReadModelEventHandler, error) {
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
		PeriodStudentAdded,
		PeriodStudentRemoved,
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
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	period_id, _ := scope[period.PeriodIDField].(string)
	if period_id == "" {
		return fmt.Errorf("no period id provided for read model event")
	}
	student_id, _ := scope[student.StudentIDField].(string)
	if student_id == "" {
		return fmt.Errorf("no student id provided for read model event")
	}

	switch resolved.Event.EventType {
	case PeriodStudentAdded:
		periodID, _ := data[period.PeriodIDField].(string)
		studentID, _ := data[student.StudentIDField].(string)
		if err := h.readModel.AddStudentToPeriod(ctx, PeriodStudentAddedProjection{
			Position:  resolved.Position,
			PeriodID:  periodID,
			StudentID: studentID,
			AddedAt:   parseTime(data[PeriodStudentAddedAtField]),
		}); err != nil {
			return err
		}
	case PeriodStudentRemoved:
		periodID, _ := data[period.PeriodIDField].(string)
		studentID, _ := data[student.StudentIDField].(string)
		if err := h.readModel.RemoveStudentFromPeriod(ctx, PeriodStudentRemovedProjection{
			Position:  resolved.Position,
			PeriodID:  periodID,
			StudentID: studentID,
			RemovedAt: parseTime(data[PeriodStudentAddedAtField]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period student read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update?
	return h.publisher.Publish(ctx, Channel(period_id), "period student read model")
}
