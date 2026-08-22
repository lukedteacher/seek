package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/educators/models"
)

const EducatorReadModelEventHandlerName = "educator_read_model_event_handler"

type EducatorReadModelReader interface {
	GetByID(ctx context.Context, educatorID string) (*models.Educator, error)
	GetByUsername(ctx context.Context, username string) (*models.Educator, error)
	List(ctx context.Context) ([]models.Educator, error)
}

type EducatorReadModelWriter interface {
	Create(ctx context.Context, event EducatorCreatedProjection) error
	Update(ctx context.Context, event EducatorUpdatedProjection) error
	Archive(ctx context.Context, event EducatorArchivedProjection) error
	Delete(ctx context.Context, event EducatorDeletedProjection) error
}

type EducatorCreatedProjection struct {
	Position eventstore.Position
	EducatorState
}

type EducatorUpdatedProjection struct {
	Position eventstore.Position
	EducatorState
}

type EducatorArchivedProjection struct {
	Position   eventstore.Position
	ID         string
	ArchivedAt time.Time
}

type EducatorDeletedProjection struct {
	Position  eventstore.Position
	ID        string
	DeletedAt time.Time
}

type EducatorReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel EducatorReadModelWriter
	publisher eventstore.Publisher
}

func NewReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel EducatorReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*EducatorReadModelEventHandler,
	error,
) {
	handler := &EducatorReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            EducatorReadModelEventHandlerName,
		Query:           EducatorReadModelEventHandlerQuery(),
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

func (h *EducatorReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *EducatorReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func EducatorReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventEducatorCreated,
		EventEducatorUpdated,
		EventEducatorArchived,
		EventEducatorDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *EducatorReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	println("educator read model handle")
	data := resolved.Event.Data
	rawData := resolved.Event.RawData
	scope := eventstore.Scope(data)
	educatorID, _ := scope[FieldEducatorID].(string)
	switch resolved.Event.EventType {
	case EventEducatorCreated:
		println("handle educator created")
		var event EducatorCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			return err
		}
		if err := h.readModel.Create(ctx, EducatorCreatedProjection{
			Position:      resolved.Position,
			EducatorState: event.EducatorState,
		}); err != nil {
			return err
		}
	case EventEducatorUpdated:
		println("handle educator updated")
		var event EducatorUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			return err
		}
		if err := h.readModel.Update(ctx, EducatorUpdatedProjection{
			Position:      resolved.Position,
			EducatorState: event.EducatorState,
		}); err != nil {
			return err
		}
	case EventEducatorArchived:
		if err := h.readModel.Archive(ctx, EducatorArchivedProjection{
			Position:   resolved.Position,
			ID:         educatorID,
			ArchivedAt: parseTime(data[FieldEducatorArchivedAt]),
		}); err != nil {
			return err
		}
	case EventEducatorDeleted:
		if err := h.readModel.Delete(ctx, EducatorDeletedProjection{
			Position:  resolved.Position,
			ID:        educatorID,
			DeletedAt: parseTime(data[FieldEducatorDeletedAt]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled educator read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, educator.Channel(educatorID).. etc)
	return h.publisher.Publish(
		ctx,
		Channel(educatorID),
		map[string]string{"educatorID": educatorID, "type": resolved.Event.EventType.String()},
	)
}

func stringPtr(value string) *string {
	return &value
}

func unmarshalEvent(data map[string]any, target any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}
