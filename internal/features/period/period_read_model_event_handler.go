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
	UpdatePeriod(ctx context.Context, event PeriodUpdatedProjection) error
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

type PeriodUpdatedProjection struct {
	Position  eventstore.Position
	Id        string
	Title     string
	StartTime string
	Duration  int64
	Days      int64
	UpdatedAt time.Time
}

type PeriodDeletedProjection struct {
	Position  eventstore.Position
	Id        string
	DeletedAt time.Time
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
	eventTypes := []string{
		PeriodCreated,
		PeriodUpdated,
		PeriodDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	periodID, _ := scope[PeriodIDField].(string)
	if periodID == "" {
		return fmt.Errorf("no id provided for read model event")
	}

	switch resolved.Event.EventType {
	case PeriodCreated:
		periodCreated := PeriodCreatedProjection{
			Id:        periodID,
			Title:     data[PeriodTitleField].(string),
			StartTime: data[PeriodStartTimeField].(string),
			Duration:  int64(data[PeriodDurationField].(float64)),
			Days:      int64(data[PeriodDaysField].(float64)),
			CreatedAt: parseTime(data[PeriodCreatedAtField]),
		}
		if err := h.readModel.InsertCreatedPeriod(ctx, periodCreated); err != nil {
			return err
		}
	case PeriodUpdated:
		periodUpdated := PeriodUpdatedProjection{
			Id:        periodID,
			Title:     data[PeriodTitleField].(string),
			StartTime: data[PeriodStartTimeField].(string),
			Duration:  int64(data[PeriodDurationField].(float64)),
			Days:      int64(data[PeriodDaysField].(float64)),
			UpdatedAt: parseTime(data[PeriodUpdatedAtField]),
		}
		if err := h.readModel.UpdatePeriod(ctx, periodUpdated); err != nil {
			return err
		}
	case PeriodDeleted:
		periodDeleted := PeriodDeletedProjection{
			Position:  resolved.Position,
			Id:        periodID,
			DeletedAt: parseTime(data[PeriodDeletedAtField]),
		}
		if err := h.readModel.DeletePeriod(ctx, periodDeleted); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, period.Channel(periodID).. etc)
	return h.publisher.Publish(ctx, Channel(periodID), map[string]string{"periodID": periodID})
}
