package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/homerooms/models"
)

const HomeroomReadModelEventHandlerName = "homeroom_read_model_event_handler"

type HomeroomReadModelReader interface {
	Get(ctx context.Context, homeroomID string) (*models.Homeroom, error)
	GetWithIDs(ctx context.Context, homeroomID string) (*models.Homeroom, error)
	List(ctx context.Context) ([]models.Homeroom, error)
	ListWithIDs(ctx context.Context) ([]models.Homeroom, error)
}

type HomeroomReadModelWriter interface {
	CreateHomeroom(ctx context.Context, event HomeroomCreatedProjection) error
	UpdateHomeroom(ctx context.Context, event HomeroomUpdatedProjection) error
	ArchiveHomeroom(ctx context.Context, event HomeroomArchivedProjection) error
	DeleteHomeroom(ctx context.Context, event HomeroomDeletedProjection) error
	AddEducatorToHomeroom(ctx context.Context, event EducatorAddedToHomeroomProjection) error
	RemoveEducatorFromHomeroom(ctx context.Context, event EducatorRemovedFromHomeroomProjection) error
	AddStudentToHomeroom(ctx context.Context, event StudentAddedToHomeroomProjection) error
	RemoveStudentFromHomeroom(ctx context.Context, event StudentRemovedFromHomeroomProjection) error
}

type HomeroomCreatedProjection struct {
	Position  eventstore.Position
	Homeroom  models.Homeroom
	CreatedAt time.Time
}

type HomeroomUpdatedProjection struct {
	Position  eventstore.Position
	Homeroom  models.Homeroom
	UpdatedAt time.Time
}

type HomeroomArchivedProjection struct {
	Position   eventstore.Position
	HomeroomID string
	ArchivedAt time.Time
}

type HomeroomDeletedProjection struct {
	Position   eventstore.Position
	HomeroomID string
	DeletedAt  time.Time
}

type EducatorAddedToHomeroomProjection struct {
	Position   eventstore.Position
	HomeroomID string
	EducatorID string
	AddedAt    time.Time
}

type EducatorRemovedFromHomeroomProjection struct {
	Position   eventstore.Position
	HomeroomID string
	EducatorID string
	RemovedAt  time.Time
}

type StudentAddedToHomeroomProjection struct {
	Position   eventstore.Position
	HomeroomID string
	StudentID  string
	AddedAt    time.Time
}

type StudentRemovedFromHomeroomProjection struct {
	Position   eventstore.Position
	HomeroomID string
	StudentID  string
	RemovedAt  time.Time
}

type HomeroomReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel HomeroomReadModelWriter
	publisher eventstore.Publisher
}

func NewReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel HomeroomReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*HomeroomReadModelEventHandler,
	error,
) {
	handler := &HomeroomReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            HomeroomReadModelEventHandlerName,
		Query:           HomeroomReadModelEventHandlerQuery(),
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

func (h *HomeroomReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *HomeroomReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func HomeroomReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
		EventEducatorAddedToHomeroom,
		EventEducatorRemovedFromHomeroom,
		EventStudentAddedToHomeroom,
		EventStudentRemovedFromHomeroom,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *HomeroomReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	rawData := resolved.Event.RawData
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	homeroomID, _ := scope[FieldHomeroomID].(string)
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		var event HomeroomCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("homeroom rm handle create unmarshal", "err", err)
		}
		projection := HomeroomCreatedProjection{
			Homeroom: models.Homeroom{
				ID:            homeroomID,
				Title:         event.Title,
				GradesBitmask: sharedmodels.GradesBitmask(event.GradesBitmask),
				LocationID:    event.LocationID,
				Image:         event.Image,
			},
		}
		if err := h.readModel.CreateHomeroom(ctx, projection); err != nil {
			return err
		}
	case EventHomeroomUpdated:
		var event HomeroomUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("homeroom rm handle update unmarshal", "err", err)
		}
		projection := HomeroomUpdatedProjection{
			Homeroom: models.Homeroom{
				ID:            homeroomID,
				Title:         event.Title,
				GradesBitmask: sharedmodels.GradesBitmask(event.GradesBitmask),
				LocationID:    event.LocationID,
				Image:         event.Image,
			},
		}
		if err := h.readModel.UpdateHomeroom(ctx, projection); err != nil {
			return err
		}
	case EventHomeroomArchived:
		homeroomArchived := HomeroomArchivedProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			ArchivedAt: parseDBTime(data[FieldHomeroomArchivedAt].(string)),
		}
		if err := h.readModel.ArchiveHomeroom(ctx, homeroomArchived); err != nil {
			return err
		}
	case EventHomeroomDeleted:
		homeroomDeleted := HomeroomDeletedProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			DeletedAt:  parseTime(data[FieldHomeroomDeletedAt]),
		}
		if err := h.readModel.DeleteHomeroom(ctx, homeroomDeleted); err != nil {
			return err
		}
	case EventEducatorAddedToHomeroom:
		educatorID, _ := scope[FieldHomeroomEducatorEducatorID].(string)
		projection := EducatorAddedToHomeroomProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			EducatorID: educatorID,
		}
		if err := h.readModel.AddEducatorToHomeroom(ctx, projection); err != nil {
			return err
		}
	case EventEducatorRemovedFromHomeroom:
		educatorID, _ := scope[FieldHomeroomEducatorEducatorID].(string)
		projection := EducatorRemovedFromHomeroomProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			EducatorID: educatorID,
		}
		if err := h.readModel.RemoveEducatorFromHomeroom(ctx, projection); err != nil {
			return err
		}
	case EventStudentAddedToHomeroom:
		studentID, _ := scope[FieldHomeroomStudentStudentID].(string)
		projection := StudentAddedToHomeroomProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			StudentID:  studentID,
		}
		if err := h.readModel.AddStudentToHomeroom(ctx, projection); err != nil {
			return err
		}
	case EventStudentRemovedFromHomeroom:
		studentID, _ := scope[FieldHomeroomStudentStudentID].(string)
		projection := StudentRemovedFromHomeroomProjection{
			Position:   resolved.Position,
			HomeroomID: homeroomID,
			StudentID:  studentID,
		}
		if err := h.readModel.RemoveStudentFromHomeroom(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled homeroom read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, homeroom.Channel(homeroomID).. etc)
	return h.publisher.Publish(ctx, Channel(homeroomID), map[string]string{"homeroomID": homeroomID})
}
