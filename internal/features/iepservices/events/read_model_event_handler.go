package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/iepservices/models"
	se "seek/internal/features/students/events"
)

const IEPServiceReadModelEventHandlerName = "iep_service_read_model_event_handler"

type IEPServiceReadModelReader interface {
	Get(ctx context.Context, iepServiceID string) (*models.IEPService, error)
	List(ctx context.Context) ([]models.IEPService, error)
	ListServicesForIEP(ctx context.Context, studentID string) ([]models.IEPService, error)
}

type IEPServiceReadModelWriter interface {
	AddServiceToIEP(ctx context.Context, event ServiceAddedToIEPProjection) error
	UpdateIEPService(ctx context.Context, event IEPServiceUpdatedProjection) error
	DeleteIEPService(ctx context.Context, event IEPServiceDeletedProjection) error
}

type ServiceAddedToIEPProjection struct {
	Position        eventstore.Position
	IEPServiceID    string
	IEPID           string
	StudentID       string
	ServiceName     string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	LocationID      string
	StartDate       string
	EndDate         string
	ProviderID      string
	CreatedAt       time.Time
}

type IEPServiceUpdatedProjection struct {
	Position        eventstore.Position
	IEPServiceID    string
	IEPID           string
	StudentID       string
	ServiceName     string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	LocationID      string
	StartDate       string
	EndDate         string
	ProviderID      string
	UpdatedAt       time.Time
}

type IEPServiceDeletedProjection struct {
	Position     eventstore.Position
	IEPServiceID string
	DeletedAt    time.Time
}

type IEPServiceReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel IEPServiceReadModelWriter
	publisher eventstore.Publisher
}

func NewIEPServiceReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel IEPServiceReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*IEPServiceReadModelEventHandler, error) {
	handler := &IEPServiceReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            IEPServiceReadModelEventHandlerName,
		Query:           IEPServiceReadModelEventHandlerQuery(),
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

func (h *IEPServiceReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *IEPServiceReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func IEPServiceReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventServiceAddedToIEP,
		EventIEPServiceUpdated,
		EventIEPServiceDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *IEPServiceReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	eventID, _ := scope[FieldServiceAddedToIEPEventID].(string)
	iepID, _ := scope[FieldIEPServiceIEPID].(string)
	studentID, _ := scope[FieldIEPServiceStudentID].(string)
	switch resolved.Event.EventType {
	case EventServiceAddedToIEP:
		projection := ServiceAddedToIEPProjection{
			IEPServiceID:    eventID,
			IEPID:           iepID,
			StudentID:       studentID,
			ServiceName:     data[FieldIEPServiceServiceName].(string),
			ServiceType:     data[FieldIEPServiceServiceType].(string),
			IndirectMinutes: int(data[FieldIEPServiceIndirectMinutes].(float64)),
			DirectMinutes:   int(data[FieldIEPServiceDirectMinutes].(float64)),
			FrequencyCount:  int(data[FieldIEPServiceFrequencyCount].(float64)),
			FrequencyType:   data[FieldIEPServiceFrequencyType].(string),
			LocationID:      data[FieldIEPServiceLocationID].(string),
			StartDate:       data[FieldIEPServiceStartDate].(string),
			EndDate:         data[FieldIEPServiceEndDate].(string),
			ProviderID:      data[FieldIEPServiceProviderID].(string),
			CreatedAt:       parseTime(data[FieldIEPServiceAddedAt]),
		}
		if err := h.readModel.AddServiceToIEP(ctx, projection); err != nil {
			return err
		}
	case EventIEPServiceUpdated:
		projection := IEPServiceUpdatedProjection{
			IEPServiceID:    eventID,
			IEPID:           iepID,
			StudentID:       studentID,
			ServiceName:     data[FieldIEPServiceServiceName].(string),
			ServiceType:     data[FieldIEPServiceServiceType].(string),
			IndirectMinutes: int(data[FieldIEPServiceIndirectMinutes].(float64)),
			DirectMinutes:   int(data[FieldIEPServiceDirectMinutes].(float64)),
			FrequencyCount:  int(data[FieldIEPServiceFrequencyCount].(float64)),
			FrequencyType:   data[FieldIEPServiceFrequencyType].(string),
			LocationID:      data[FieldIEPServiceLocationID].(string),
			StartDate:       data[FieldIEPServiceStartDate].(string),
			EndDate:         data[FieldIEPServiceEndDate].(string),
			ProviderID:      data[FieldIEPServiceProviderID].(string),
			UpdatedAt:       parseTime(data[FieldIEPServiceUpdatedAt]),
		}
		if err := h.readModel.UpdateIEPService(ctx, projection); err != nil {
			return err
		}
	case EventIEPServiceDeleted:
		projection := IEPServiceDeletedProjection{
			Position:     resolved.Position,
			IEPServiceID: eventID,
			DeletedAt:    parseTime(data[FieldIEPServiceDeletedAt]),
		}
		if err := h.readModel.DeleteIEPService(ctx, projection); err != nil {
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
