package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"
)

const PeriodReadModelEventHandlerName = "period_read_model_event_handler"

type PeriodReadModelReader interface {
	Get(ctx context.Context, periodID string) (*models.Period, error)
	GetWithIDs(ctx context.Context, periodID string) (*models.Period, error)
	List(ctx context.Context) ([]models.Period, error)
	ListPeriodsForStudent(ctx context.Context, studentID string) ([]models.Period, error)
	ListPeriodsForEducator(ctx context.Context, educatorID string) ([]models.Period, error)
}

type PeriodReadModelWriter interface {
	CreatePeriod(ctx context.Context, event PeriodCreatedProjection) error
	UpdatePeriod(ctx context.Context, event PeriodUpdatedProjection) error
	ArchivePeriod(ctx context.Context, event PeriodArchivedProjection) error
	DeletePeriod(ctx context.Context, event PeriodDeletedProjection) error
}

type PeriodCreatedProjection struct {
	Position    eventstore.Position
	PeriodID    string
	Title       string
	ServiceType sharedmodels.ServiceType
	StartTime   sharedmodels.TimeOnly
	Duration    int
	DaysBitmask sharedmodels.DaysBitmask
	CreatedAt   time.Time
}

type PeriodUpdatedProjection struct {
	Position    eventstore.Position
	PeriodID    string
	Title       string
	ServiceType sharedmodels.ServiceType
	StartTime   sharedmodels.TimeOnly
	Duration    int
	DaysBitmask sharedmodels.DaysBitmask
	UpdatedAt   time.Time
}

type PeriodArchivedProjection struct {
	Position   eventstore.Position
	PeriodID   string
	ArchivedAt time.Time
}

type PeriodDeletedProjection struct {
	Position  eventstore.Position
	PeriodID  string
	DeletedAt time.Time
}

type PeriodReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel PeriodReadModelWriter
	publisher eventstore.Publisher
}

func NewPeriodReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel PeriodReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*PeriodReadModelEventHandler,
	error,
) {
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
		PeriodArchived,
		PeriodDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *PeriodReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	periodID, _ := scope[FieldPeriodID].(string)
	// this is to prevent errors where the period ID isn't present or read correctly
	if periodID == "" {
		return fmt.Errorf("no id provided for period read model event")
	}

	switch resolved.Event.EventType {
	case PeriodCreated:
		periodCreated := PeriodCreatedProjection{
			PeriodID:    periodID,
			Title:       data[FieldPeriodTitle].(string),
			ServiceType: sharedmodels.ServiceType(data[FieldPeriodServiceType].(string)),
			StartTime:   parseDBTimeOnly(data[FieldPeriodStartTime].(string)),
			Duration:    int(data[FieldPeriodDuration].(float64)),
			DaysBitmask: sharedmodels.DaysBitmask(data[FieldPeriodDaysBitmask].(float64)),
			CreatedAt:   parseDBTime(data[FieldPeriodCreatedAt].(string)),
		}
		if err := h.readModel.CreatePeriod(ctx, periodCreated); err != nil {
			return err
		}
	case PeriodUpdated:
		periodUpdated := PeriodUpdatedProjection{
			PeriodID:    periodID,
			Title:       data[FieldPeriodTitle].(string),
			ServiceType: sharedmodels.ServiceType(data[FieldPeriodServiceType].(string)),
			StartTime:   parseDBTimeOnly(data[FieldPeriodStartTime].(string)),
			Duration:    int(data[FieldPeriodDuration].(float64)),
			DaysBitmask: sharedmodels.DaysBitmask(data[FieldPeriodDaysBitmask].(float64)),
			UpdatedAt:   parseDBTime(data[FieldPeriodUpdatedAt].(string)),
		}
		if err := h.readModel.UpdatePeriod(ctx, periodUpdated); err != nil {
			return err
		}
	case PeriodArchived:
		periodArchived := PeriodArchivedProjection{
			Position:   resolved.Position,
			PeriodID:   periodID,
			ArchivedAt: parseDBTime(data[FieldPeriodArchivedAt].(string)),
		}
		if err := h.readModel.ArchivePeriod(ctx, periodArchived); err != nil {
			return err
		}
	case PeriodDeleted:
		periodDeleted := PeriodDeletedProjection{
			Position:  resolved.Position,
			PeriodID:  periodID,
			DeletedAt: parseTime(data[FieldPeriodDeletedAt]),
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
