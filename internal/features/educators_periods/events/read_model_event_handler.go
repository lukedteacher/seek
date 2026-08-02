package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	educator "seek/internal/features/educators/events"
	period "seek/internal/features/periods/events"
)

const PeriodEducatorReadModelEventHandlerName = "period_educator_read_model_event_handler"

type EducatorAddedToPeriodProjection struct {
	Position   eventstore.Position
	PeriodID   string
	EducatorID string
	AddedAt    time.Time
}

type EducatorRemovedFromPeriodProjection struct {
	Position   eventstore.Position
	PeriodID   string
	EducatorID string
	RemovedAt  time.Time
}

type PeriodEducatorReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel PeriodEducatorReadModelWriter
	publisher eventstore.Publisher
}

func NewPeriodEducatorReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel PeriodEducatorReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*PeriodEducatorReadModelEventHandler,
	error,
) {
	handler := &PeriodEducatorReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            PeriodEducatorReadModelEventHandlerName,
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

func (h *PeriodEducatorReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *PeriodEducatorReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func ScheduleReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		EducatorAddedToPeriod,
		EducatorRemovedFromPeriod,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodEducatorReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	var periodID, educatorID string
	println("**********")
	switch resolved.Event.EventType {
	case EducatorAddedToPeriod:
		var event EducatorAddedToPeriodEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		periodID = event.PeriodID
		educatorID = event.EducatorID
		if err := h.readModel.AddEducatorToPeriod(ctx, EducatorAddedToPeriodProjection{
			Position:   resolved.Position,
			PeriodID:   event.PeriodID,
			EducatorID: event.EducatorID,
			AddedAt:    event.AddedAt,
		}); err != nil {
			return err
		}
	case EducatorRemovedFromPeriod:
		var event EducatorRemovedFromPeriodEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		periodID = event.PeriodID
		educatorID = event.EducatorID
		if err := h.readModel.RemoveEducatorFromPeriod(ctx, EducatorRemovedFromPeriodProjection{
			Position:   resolved.Position,
			PeriodID:   event.PeriodID,
			EducatorID: event.EducatorID,
			RemovedAt:  event.RemovedAt,
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period educator read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	_ = h.publisher.Publish(ctx, period.Channel(periodID), "period educator read model update")
	_ = h.publisher.Publish(ctx, educator.Channel(educatorID), "period educator read model update")
	return nil
}
