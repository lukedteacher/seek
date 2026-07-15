package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	schedule "seek/internal/features/schedules/events"
)

const PeriodScheduleReadModelEventHandlerName = "period_schedule_read_model_event_handler"

type PeriodScheduleAddedProjection struct {
	Position   eventstore.Position
	PeriodID   string
	ScheduleID string
	AddedAt    time.Time
}

type PeriodScheduleRemovedProjection struct {
	Position   eventstore.Position
	PeriodID   string
	ScheduleID string
	RemovedAt  time.Time
}

type PeriodScheduleReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel PeriodScheduleReadModelWriter
	publisher eventstore.Publisher
}

func NewPeriodScheduleReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel PeriodScheduleReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*PeriodScheduleReadModelEventHandler,
	error,
) {
	handler := &PeriodScheduleReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            PeriodScheduleReadModelEventHandlerName,
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

func (h *PeriodScheduleReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *PeriodScheduleReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func ScheduleReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		PeriodScheduleAdded,
		PeriodScheduleRemoved,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodScheduleReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	period_id, _ := scope[period.PeriodIDField].(string)
	if period_id == "" {
		return fmt.Errorf("no period id provided for read model event")
	}
	schedule_id, _ := scope[schedule.ScheduleIDField].(string)
	if schedule_id == "" {
		return fmt.Errorf("no schedule id provided for read model event")
	}

	switch resolved.Event.EventType {
	case PeriodScheduleAdded:
		periodID, _ := data[period.PeriodIDField].(string)
		scheduleID, _ := data[schedule.ScheduleIDField].(string)
		if err := h.readModel.AddPeriodToSchedule(ctx, PeriodScheduleAddedProjection{
			Position:   resolved.Position,
			PeriodID:   periodID,
			ScheduleID: scheduleID,
			AddedAt:    parseTime(data[PeriodScheduleAddedAtField]),
		}); err != nil {
			return err
		}
	case PeriodScheduleRemoved:
		periodID, _ := data[period.PeriodIDField].(string)
		scheduleID, _ := data[schedule.ScheduleIDField].(string)
		if err := h.readModel.RemovePeriodFromSchedule(ctx, PeriodScheduleRemovedProjection{
			Position:   resolved.Position,
			PeriodID:   periodID,
			ScheduleID: scheduleID,
			RemovedAt:  parseTime(data[PeriodScheduleAddedAtField]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period schedule read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update?
	return h.publisher.Publish(ctx, Channel(period_id), "period schedule read model")
}
