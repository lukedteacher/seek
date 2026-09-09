package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/services/models"
	se "seek/internal/features/students/events"
)

const ServiceReadModelEventHandlerName = "iep_service_read_model_event_handler"

type ServiceReadModelReader interface {
	Get(ctx context.Context, serviceID string) (*models.Service, error)
	List(ctx context.Context) ([]models.Service, error)
	ListServicesForIEP(ctx context.Context, studentID string) ([]models.Service, error)
}

type ServiceReadModelWriter interface {
	AddServiceToIEP(ctx context.Context, event ServiceAddedToIEPProjection) error
	UpdateService(ctx context.Context, event ServiceUpdatedProjection) error
	DeleteService(ctx context.Context, event ServiceDeletedProjection) error
}

type ServiceAddedToIEPProjection struct {
	Position eventstore.Position
	Service  models.Service
}

type ServiceUpdatedProjection struct {
	Position eventstore.Position
	Service  models.Service
}

type ServiceDeletedProjection struct {
	Position  eventstore.Position
	ServiceID string
	DeletedAt time.Time
}

type ServiceReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel ServiceReadModelWriter
	publisher eventstore.Publisher
}

func NewServiceReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel ServiceReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*ServiceReadModelEventHandler, error) {
	handler := &ServiceReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            ServiceReadModelEventHandlerName,
		Query:           ServiceReadModelEventHandlerQuery(),
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

func (h *ServiceReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *ServiceReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func ServiceReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventServiceAddedToIEP,
		EventServiceUpdated,
		EventServiceDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *ServiceReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	rawData := resolved.Event.RawData
	scope := eventstore.Scope(data)
	eventID, _ := scope[FieldServiceAddedToIEPEventID].(string)
	studentID, _ := scope[FieldServiceStudentID].(string)
	switch resolved.Event.EventType {
	case EventServiceAddedToIEP:
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			return err
		}
		model := NewModelFromFlat(flat)
		projection := ServiceAddedToIEPProjection{
			Position: resolved.Position,
			Service:  model,
		}
		if err := h.readModel.AddServiceToIEP(ctx, projection); err != nil {
			return err
		}
	case EventServiceUpdated:
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			return err
		}
		model := NewModelFromFlat(flat)
		projection := ServiceUpdatedProjection{
			Position: resolved.Position,
			Service:  model,
		}
		if err := h.readModel.UpdateService(ctx, projection); err != nil {
			return err
		}
	case EventServiceDeleted:
		projection := ServiceDeletedProjection{
			Position:  resolved.Position,
			ServiceID: eventID,
			DeletedAt: parseTime(data[FieldServiceDeletedAt]),
		}
		if err := h.readModel.DeleteService(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, period.Channel(periodID).. etc)
	_ = h.publisher.Publish(ctx, Channel(eventID), "iep service read model update")
	_ = h.publisher.Publish(ctx, Channel(eventID), "iep service read model update")
	_ = h.publisher.Publish(ctx, se.Channel(studentID), "student read model update")
	return nil
}
