package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"seek/internal/eventstore"
	"seek/internal/features/ieps/models"
	se "seek/internal/features/students/events"
)

const IEPReadModelEventHandlerName = "iep_read_model_event_handler"

type StudentIEPReadModelReader interface {
	Get(ctx context.Context, studentIEPID string) (*models.IEP, error)
	List(ctx context.Context) ([]models.IEP, error)
	ListIEPsForStudent(ctx context.Context, studentID string) ([]models.IEP, error)
}

type StudentIEPReadModelWriter interface {
	AddIEPToStudent(ctx context.Context, event IEPAddedToStudentProjection) error
	UpdateIEP(ctx context.Context, event IEPUpdatedProjection) error
	DeleteIEP(ctx context.Context, event IEPDeletedProjection) error
}

type IEPAddedToStudentProjection struct {
	Position eventstore.Position
	IEPState
}

type IEPUpdatedProjection struct {
	Position eventstore.Position
	IEPState
}

type IEPArchivedProjection struct {
	Position eventstore.Position
	IEPState
}

type IEPDeletedProjection struct {
	Position eventstore.Position
	IEPState
}

type IEPReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel StudentIEPReadModelWriter
	publisher eventstore.Publisher
}

func NewIEPReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel StudentIEPReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*IEPReadModelEventHandler,
	error,
) {
	handler := &IEPReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            IEPReadModelEventHandlerName,
		Query:           IEPReadModelEventHandlerQuery(),
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

func (h *IEPReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *IEPReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func IEPReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventIEPAddedToStudent,
		EventIEPRemovedFromStudent,
		EventIEPUpdated,
		EventIEPArchived,
		EventIEPDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *IEPReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	rawData := resolved.Event.RawData
	scope := eventstore.Scope(data)
	iepID, _ := scope[FieldIEPID].(string)
	studentID, _ := scope[FieldIEPStudentID].(string)
	switch resolved.Event.EventType {
	case EventIEPAddedToStudent:
		var event IEPAddedToStudentEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			return err
		}
		projection := IEPAddedToStudentProjection{
			Position: resolved.Position,
			IEPState: event.IEPState,
		}
		if err := h.readModel.AddIEPToStudent(ctx, projection); err != nil {
			return err
		}
	case EventIEPUpdated:
		var event IEPUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep read model handle update unmarshal", "err", err)
		}
		projection := IEPUpdatedProjection{
			Position: resolved.Position,
			IEPState: event.IEPState,
		}
		if err := h.readModel.UpdateIEP(ctx, projection); err != nil {
			return err
		}
	case EventIEPDeleted:
		var event IEPDeletedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep read model handle delete unmarshal", "err", err)
		}
		projection := IEPDeletedProjection{
			Position: resolved.Position,
			IEPState: event.IEPState,
		}
		if err := h.readModel.DeleteIEP(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, iep.Channel(iepID) ...etc.)
	_ = h.publisher.Publish(ctx, Channel(iepID), "iep read model update")
	_ = h.publisher.Publish(ctx, se.Channel(studentID), "student read model update")
	return nil
}
