package schedule

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
)

const ScheduleReadModelEventHandlerName = "schedule_read_model_event_handler"

type ScheduleReadModelReader interface {
	Get(ctx context.Context, scheduleID string) (*models.Schedule, error)
	List(ctx context.Context) ([]models.Schedule, error)
	ListSchedulePeriodIDs(ctx context.Context, scheduleID string) ([]string, error)
	UpdateSchedulePeriods(ctx context.Context, currentPeriodIDs, proposedPeriodIDs []string) error
}

type ScheduleReadModelWriter interface {
	InsertCreatedSchedule(ctx context.Context, event ScheduleCreatedProjection) error
	UpdateSchedule(ctx context.Context, event ScheduleUpdatedProjection) error
	AddSchedulePeriod(ctx context.Context, event SchedulePeriodAddedProjection) error
	RemoveSchedulePeriod(ctx context.Context, event SchedulePeriodRemovedProjection) error
	DeleteSchedule(ctx context.Context, event ScheduleDeletedProjection) error
}

type ScheduleCreatedProjection struct {
	Position   eventstore.Position
	ScheduleID string
	Title      string
	TeacherId  string
	CreatedAt  time.Time
}

type ScheduleUpdatedProjection struct {
	Position   eventstore.Position
	ScheduleID string
	Title      string
	TeacherId  string
	UpdatedAt  time.Time
}

type SchedulePeriodAddedProjection struct {
	Position   eventstore.Position
	ScheduleID string
	PeriodID   string
	AddedAt    time.Time
}

type SchedulePeriodRemovedProjection struct {
	Position   eventstore.Position
	ScheduleID string
	PeriodID   string
	RemovedAt  time.Time
}

type ScheduleDeletedProjection struct {
	Position   eventstore.Position
	ScheduleID string
	DeletedAt  time.Time
}

type ScheduleReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel ScheduleReadModelWriter
	publisher eventstore.Publisher
}

func NewScheduleReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel ScheduleReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*ScheduleReadModelEventHandler, error) {
	handler := &ScheduleReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            ScheduleReadModelEventHandlerName,
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

func (h *ScheduleReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *ScheduleReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func ScheduleReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		ScheduleCreated,
		SchedulePeriodAdded,
		SchedulePeriodRemoved,
		ScheduleUpdated,
		ScheduleDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *ScheduleReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	id, _ := scope["id"].(string)
	if id == "" {
		return fmt.Errorf("no id provided for read model event")
	}

	switch resolved.Event.EventType {
	case ScheduleCreated:
		title, _ := data["title"].(string)
		teacherId, _ := data["teacherID"].(string)
		if err := h.readModel.InsertCreatedSchedule(ctx, ScheduleCreatedProjection{
			Position:   resolved.Position,
			ScheduleID: id,
			Title:      title,
			TeacherId:  teacherId,
			CreatedAt:  parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case ScheduleUpdated:
		title, _ := data["title"].(string)
		teacherId, _ := data["teacherID"].(string)
		if err := h.readModel.UpdateSchedule(ctx, ScheduleUpdatedProjection{
			Position:   resolved.Position,
			ScheduleID: id,
			Title:      title,
			TeacherId:  teacherId,
			UpdatedAt:  parseTime(data["updatedAt"]),
		}); err != nil {
			return err
		}
	case SchedulePeriodAdded:
		scheduleID, _ := data["scheduleID"].(string)
		periodID, _ := data["periodID"].(string)
		if err := h.readModel.AddSchedulePeriod(ctx, SchedulePeriodAddedProjection{
			Position:                resolved.Position,
			ScheduleID:              scheduleID,
			PeriodID:                periodID,
			AddedAt: parseTime(data["periodAddedAt"]),
		}); err != nil {
			return err
		}
	case SchedulePeriodRemoved:
		scheduleID, _ := data["scheduleID"].(string)
		periodID, _ := data["periodID"].(string)
		println("eh: ", data["periodRemovedAt"].(string))
		if err := h.readModel.RemoveSchedulePeriod(ctx, SchedulePeriodRemovedProjection{
			Position:                resolved.Position,
			ScheduleID:              scheduleID,
			PeriodID:                periodID,
			RemovedAt: parseTime(data["periodRemovedAt"]),
		}); err != nil {
			return err
		}
	case ScheduleDeleted:
		if err := h.readModel.DeleteSchedule(ctx, ScheduleDeletedProjection{
			Position:   resolved.Position,
			ScheduleID: id,
			DeletedAt:  parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled schedule read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	return h.publisher.Publish(ctx, Channel("idk"), "test")
}
