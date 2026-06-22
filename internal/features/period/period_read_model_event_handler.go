package period

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
)

const PeriodReadModelEventHandlerName = "period_read_model_event_handler"

type PeriodReadModelReader interface {
	Get(ctx context.Context, periodID string) (*models.Period, error)
	List(ctx context.Context) ([]models.Period, error)
}

type PeriodReadModelWriter interface {
	InsertCreatedPeriod(ctx context.Context, event PeriodCreatedProjection) error
	RenamePeriod(ctx context.Context, event PeriodRenamedProjection) error
	DeletePeriod(ctx context.Context, event PeriodDeletedProjection) error
}

type PeriodCreatedProjection struct {
	Position  eventstore.Position
	Id        string
	Title     string
	StartTime string
	Duration  int64
	Days      int64
	CreatedAt time.Time
}

type PeriodRenamedProjection struct {
	Position  eventstore.Position
	Id        string
	Title     string
	RenamedAt time.Time
}

type PeriodDeletedProjection struct {
	Position	eventstore.Position
	Id				string
	DeletedAt	time.Time
}

type PeriodReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel PeriodReadModelWriter
	publisher eventstore.Publisher
}

func NewPeriodReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel PeriodReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*PeriodReadModelEventHandler, error) {
	handler := &PeriodReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            PeriodReadModelEventHandlerName,
		Query:           PeriodReadModelEventHandlerQuery(),
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

func (h *PeriodReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *PeriodReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func PeriodReadModelEventHandlerQuery() eventstore.Query {
	criteria := make([]eventstore.Criterion, 0, 5)
	for _, eventType := range []string{PeriodCreated, PeriodRenamed, PeriodDeleted} {
		criteria = append(criteria, eventstore.Criterion{Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}}})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	println("period read model event handler")
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	id, _ := scope["id"].(string)

	switch resolved.Event.EventType {
	case PeriodCreated:
		title, _ := data["first_name"].(string)
		startTime, _ := data["start_time"].(string)
		duration, ok := data["duration"].(int64)
    if !ok {
      if durationFloat, ok := data["duration"].(float64); ok {
        duration = int64(durationFloat)
      } else {
				println("invalid duration")
        duration = 0
      }
    }
		days, ok := data["days"].(int64)
    if !ok {
      if daysFloat, ok := data["days"].(float64); ok {
        days = int64(daysFloat)
      } else {
				println("invalid days")
        days = 0
      }
    }
		if err := h.readModel.InsertCreatedPeriod(ctx, PeriodCreatedProjection{
			Position:  resolved.Position,
			Id:        id,
			Title:     title,
			StartTime: startTime,
			Duration:  duration,
			Days:      days,
			CreatedAt: parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case PeriodRenamed:
		title, _ := data["first_name"].(string)
		if err := h.readModel.RenamePeriod(ctx, PeriodRenamedProjection{
			Position:  resolved.Position,
			Id:        id,
			Title:     title,
			RenamedAt: parseTime(data["renamedAt"]),
		}); err != nil {
			return err
		}
	case PeriodDeleted:
		if err := h.readModel.DeletePeriod(ctx, PeriodDeletedProjection{
			Position:  resolved.Position,
			Id:        id,
			DeletedAt: parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// TODO not sure if this is the fix
	return nil
}
