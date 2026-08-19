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
	Position   eventstore.Position
	ID         string
	GivenName  string
	ChosenName string
	FamilyName string
	Email      string
	Username   string
	Roles      []string
	CreatedAt  time.Time
}

type EducatorUpdatedProjection struct {
	Position   eventstore.Position
	ID         string
	GivenName  string
	ChosenName string
	FamilyName string
	Email      string
	Username   string
	Roles      []string
	UpdatedAt  time.Time
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
	eventTypes := []string{
		EducatorCreated,
		EducatorUpdated,
		EducatorArchived,
		EducatorDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *EducatorReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	educatorCreatedEventID, _ := scope[FieldEducatorCreatedEventID].(string)
	switch resolved.Event.EventType {
	case EducatorCreated:
		var event EducatorCreatedEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		if err := h.readModel.Create(ctx, EducatorCreatedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			GivenName:  event.GivenName,
			ChosenName: event.ChosenName,
			FamilyName: event.FamilyName,
			Email:      event.Email,
			Username:   event.Username,
			Roles:      event.Roles,
			CreatedAt:  event.CreatedAt,
		}); err != nil {
			return err
		}
	case EducatorUpdated:
		var event EducatorUpdatedEvent
		if err := json.Unmarshal([]byte(resolved.Event.RawData), &event); err != nil {
			return err
		}
		if err := h.readModel.Update(ctx, EducatorUpdatedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			GivenName:  event.GivenName,
			ChosenName: event.ChosenName,
			FamilyName: event.FamilyName,
			Email:      event.Email,
			Username:   event.Username,
			Roles:      event.Roles,
			UpdatedAt:  parseTime(data[FieldEducatorUpdatedAt]),
		}); err != nil {
			return err
		}
	case EducatorArchived:
		if err := h.readModel.Archive(ctx, EducatorArchivedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			ArchivedAt: parseTime(data[FieldEducatorArchivedAt]),
		}); err != nil {
			return err
		}
	case EducatorDeleted:
		if err := h.readModel.Delete(ctx, EducatorDeletedProjection{
			Position:  resolved.Position,
			ID:        educatorCreatedEventID,
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
		Channel(educatorCreatedEventID),
		map[string]string{"educatorID": educatorCreatedEventID},
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
