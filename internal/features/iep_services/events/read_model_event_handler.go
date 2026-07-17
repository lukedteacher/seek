package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/iep_services/models"
	se "seek/internal/features/students/events"
)

const IEPServiceReadModelEventHandlerName = "student_service_read_model_event_handler"

type IEPServiceReadModelReader interface {
	Get(ctx context.Context, iepServiceID string) (*models.IEPService, error)
	List(ctx context.Context) ([]models.IEPService, error)
}

type IEPServiceReadModelWriter interface {
	CreateIEPService(ctx context.Context, event IEPServiceCreatedProjection) error
	UpdateIEPService(ctx context.Context, event IEPServiceUpdatedProjection) error
	DeleteIEPService(ctx context.Context, event IEPServiceDeletedProjection) error
}

type IEPServiceCreatedProjection struct {
	Position        eventstore.Position
	IEPServiceID    string
	StudentID       string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
	CreatedAt       time.Time
}

type IEPServiceUpdatedProjection struct {
	Position        eventstore.Position
	IEPServiceID    string
	StudentID       string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
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
	eventTypes := []string{
		IEPServiceCreated,
		IEPServiceUpdated,
		IEPServiceDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *IEPServiceReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	iepServiceID, _ := scope[IEPServiceIDField].(string)
	studentID, _ := scope[IEPServiceStudentIDField].(string)
	// this is to prevent errors where the period ID isn't present or read correctly
	if iepServiceID == "" {
		return fmt.Errorf("no id provided for student service read model event")
	}

	switch resolved.Event.EventType {
	case IEPServiceCreated:
		projection := IEPServiceCreatedProjection{
			IEPServiceID:    iepServiceID,
			StudentID:       studentID,
			ServiceType:     data[IEPServiceServiceTypeField].(string),
			IndirectMinutes: int(data[IEPServiceIndirectMinutesField].(float64)),
			DirectMinutes:   int(data[IEPServiceDirectMinutesField].(float64)),
			FrequencyCount:  int(data[IEPServiceFrequencyCountField].(float64)),
			FrequencyType:   data[IEPServiceFrequencyTypeField].(string),
			Location:        data[IEPServiceLocationField].(string),
			StartDate:       data[IEPServiceStartDateField].(string),
			EndDate:         data[IEPServiceEndDateField].(string),
			Provider:        data[IEPServiceProviderField].(string),
			CreatedAt:       parseTime(data[IEPServiceCreatedAtField]),
		}
		if err := h.readModel.CreateIEPService(ctx, projection); err != nil {
			return err
		}

	case IEPServiceUpdated:
		projection := IEPServiceUpdatedProjection{
			IEPServiceID:    iepServiceID,
			StudentID:       studentID,
			ServiceType:     data[IEPServiceServiceTypeField].(string),
			IndirectMinutes: int(data[IEPServiceIndirectMinutesField].(float64)),
			DirectMinutes:   int(data[IEPServiceDirectMinutesField].(float64)),
			FrequencyCount:  int(data[IEPServiceFrequencyCountField].(float64)),
			FrequencyType:   data[IEPServiceFrequencyTypeField].(string),
			Location:        data[IEPServiceLocationField].(string),
			StartDate:       data[IEPServiceStartDateField].(string),
			EndDate:         data[IEPServiceEndDateField].(string),
			Provider:        data[IEPServiceProviderField].(string),
			UpdatedAt:       parseTime(data[IEPServiceUpdatedAtField]),
		}
		if err := h.readModel.UpdateIEPService(ctx, projection); err != nil {
			return err
		}

	case IEPServiceDeleted:
		projection := IEPServiceDeletedProjection{
			Position:     resolved.Position,
			IEPServiceID: iepServiceID,
			DeletedAt:    parseTime(data[IEPServiceDeletedAtField]),
		}
		if err := h.readModel.DeleteIEPService(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, period.Channel(periodID).. etc)
	_ = h.publisher.Publish(ctx, Channel(iepServiceID), "student service read model update")
	_ = h.publisher.Publish(ctx, se.Channel(studentID), "student service read model update")
	return nil
}
