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
}

type ScheduleReadModelWriter interface {
	InsertCreatedSchedule(ctx context.Context, event ScheduleCreatedProjection) error
	UpdateSchedule(ctx context.Context, event ScheduleUpdatedProjection) error
	PeriodAddedToSchedule(ctx context.Context, event PeriodAddedToScheduleProjection) error
	DeleteSchedule(ctx context.Context, event ScheduleDeletedProjection) error
}

type ScheduleCreatedProjection struct {
	Position  eventstore.Position
	ScheduleID        string
	Title     string
	TeacherId string
	CreatedAt time.Time
}

type ScheduleUpdatedProjection struct {
	Position  eventstore.Position
	ScheduleID        string
	Title     string
	TeacherId string
	UpdatedAt time.Time
}

type PeriodAddedToScheduleProjection struct {
	Position                eventstore.Position
	ScheduleID              string
	PeriodID                string
	PeriodAddedToScheduleAt time.Time
}

type ScheduleDeletedProjection struct {
	Position  eventstore.Position
	ScheduleID        string
	DeletedAt time.Time
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
		ScheduleUpdated,
		PeriodAddedToSchedule,
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
	println(resolved.Event.Data)
	switch resolved.Event.EventType {
	case ScheduleCreated:
		println("sc case in read model")
		title, _ := data["title"].(string)
		teacherId, _ := data["teacherID"].(string)
		if err := h.readModel.InsertCreatedSchedule(ctx, ScheduleCreatedProjection{
			Position:  resolved.Position,
			ScheduleID:        id,
			Title:     title,
			TeacherId: teacherId,
			CreatedAt: parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case ScheduleUpdated:
		println("su case in read model")
		title, _ := data["title"].(string)
		teacherId, _ := data["teacherID"].(string)
		if err := h.readModel.UpdateSchedule(ctx, ScheduleUpdatedProjection{
			Position:  resolved.Position,
			ScheduleID:        id,
			Title:     title,
			TeacherId: teacherId,
			UpdatedAt: parseTime(data["updatedAt"]),
		}); err != nil {
			return err
		}
	case PeriodAddedToSchedule:
		println("pa case in read model")
		scheduleID, _ := data["scheduleID"].(string)
		println("c sid: ", scheduleID)
		periodID, _ := data["periodID"].(string)
		if err := h.readModel.PeriodAddedToSchedule(ctx, PeriodAddedToScheduleProjection{
			Position:                resolved.Position,
			ScheduleID:              scheduleID,
			PeriodID:                periodID,
			PeriodAddedToScheduleAt: parseTime(data["periodAddedToScheduleAt"]),
		}); err != nil {
			return err
		}
	case ScheduleDeleted:
		if err := h.readModel.DeleteSchedule(ctx, ScheduleDeletedProjection{
			Position:  resolved.Position,
			ScheduleID:        id,
			DeletedAt: parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled schedule read model event type %q", resolved.Event.EventType)
	}
	// TODO not sure if this is the fix
	return nil
}
