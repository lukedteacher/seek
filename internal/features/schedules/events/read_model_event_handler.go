package events

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
	CreateSchedule(ctx context.Context, event ScheduleCreatedProjection) error
	UpdateSchedule(ctx context.Context, event ScheduleUpdatedProjection) error
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

// reads all schedule events
func ScheduleReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		ScheduleCreated,
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
	scheduleID, _ := scope[ScheduleIDField].(string)
	if scheduleID == "" {
		return fmt.Errorf("no id provided for schedule read model event")
	}

	switch resolved.Event.EventType {
	case ScheduleCreated:
		title, _ := data[ScheduleTitleField].(string)
		teacherId, _ := data[ScheduleTeacherIDField].(string)
		if err := h.readModel.CreateSchedule(ctx, ScheduleCreatedProjection{
			Position:   resolved.Position,
			ScheduleID: scheduleID,
			Title:      title,
			TeacherId:  teacherId,
			CreatedAt:  parseTime(data[ScheduleCreatedAtField]),
		}); err != nil {
			return err
		}
	case ScheduleUpdated:
		title, _ := data[ScheduleTitleField].(string)
		teacherId, _ := data[ScheduleTeacherIDField].(string)
		if err := h.readModel.UpdateSchedule(ctx, ScheduleUpdatedProjection{
			Position:   resolved.Position,
			ScheduleID: scheduleID,
			Title:      title,
			TeacherId:  teacherId,
			UpdatedAt:  parseTime(data[ScheduleUpdatedAtField]),
		}); err != nil {
			return err
		}
	case ScheduleDeleted:
		if err := h.readModel.DeleteSchedule(ctx, ScheduleDeletedProjection{
			Position:   resolved.Position,
			ScheduleID: scheduleID,
			DeletedAt:  parseTime(data[ScheduleDeletedAtField]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled schedule read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, schedule.Channel(scheduleID).. etc)
	return h.publisher.Publish(
		ctx,
		Channel(scheduleID),
		map[string]string{"scheduleID": scheduleID},
	)
}
