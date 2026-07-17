package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
)

const StudentServiceReadModelEventHandlerName = "student_service_read_model_event_handler"

type StudentServiceReadModelReader interface {
	Get(ctx context.Context, serviceID string) (*models.StudentService, error)
	List(ctx context.Context) ([]models.StudentService, error)
	ListServicesForStudent(ctx context.Context, studentID string) ([]models.StudentService, error)
}

type StudentServiceReadModelWriter interface {
	CreateStudentService(ctx context.Context, event StudentServiceCreatedProjection) error
	UpdateStudentService(ctx context.Context, event StudentServiceUpdatedProjection) error
	DeleteStudentService(ctx context.Context, event StudentServiceDeletedProjection) error
}

type StudentServiceCreatedProjection struct {
	Position        eventstore.Position
	ServiceID       string
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

type StudentServiceUpdatedProjection struct {
	Position        eventstore.Position
	ServiceID       string
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

type StudentServiceDeletedProjection struct {
	Position  eventstore.Position
	ServiceID string
	DeletedAt time.Time
}

type StudentServiceReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel StudentServiceReadModelWriter
	publisher eventstore.Publisher
}

func NewStudentServiceReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel StudentServiceReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*StudentServiceReadModelEventHandler, error) {
	handler := &StudentServiceReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            StudentServiceReadModelEventHandlerName,
		Query:           StudentServiceReadModelEventHandlerQuery(),
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

func (h *StudentServiceReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *StudentServiceReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func StudentServiceReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		StudentServiceCreated,
		StudentServiceUpdated,
		StudentServiceDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *StudentServiceReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	serviceID, _ := scope[StudentServiceIDField].(string)
	studentID, _ := scope[StudentServiceStudentIDField].(string)
	// this is to prevent errors where the period ID isn't present or read correctly
	if serviceID == "" {
		return fmt.Errorf("no id provided for student service read model event")
	}

	switch resolved.Event.EventType {
	case StudentServiceCreated:
		projection := StudentServiceCreatedProjection{
			ServiceID:       serviceID,
			StudentID:       studentID,
			ServiceType:     data[StudentServiceTypeField].(string),
			IndirectMinutes: int(data[StudentServiceIndirectMinutesField].(float64)),
			DirectMinutes:   int(data[StudentServiceDirectMinutesField].(float64)),
			FrequencyCount:  int(data[StudentServiceFrequencyCountField].(float64)),
			FrequencyType:   data[StudentServiceFrequencyTypeField].(string),
			Location:        data[StudentServiceLocationField].(string),
			StartDate:       data[StudentServiceStartDateField].(string),
			EndDate:         data[StudentServiceEndDateField].(string),
			Provider:        data[StudentServiceProviderField].(string),
			CreatedAt:       parseTime(data[StudentServiceCreatedAtField]),
		}
		if err := h.readModel.CreateStudentService(ctx, projection); err != nil {
			return err
		}

	case StudentServiceUpdated:
		projection := StudentServiceUpdatedProjection{
			ServiceID:       serviceID,
			StudentID:       studentID,
			ServiceType:     data[StudentServiceTypeField].(string),
			IndirectMinutes: int(data[StudentServiceIndirectMinutesField].(float64)),
			DirectMinutes:   int(data[StudentServiceDirectMinutesField].(float64)),
			FrequencyCount:  int(data[StudentServiceFrequencyCountField].(float64)),
			FrequencyType:   data[StudentServiceFrequencyTypeField].(string),
			Location:        data[StudentServiceLocationField].(string),
			StartDate:       data[StudentServiceStartDateField].(string),
			EndDate:         data[StudentServiceEndDateField].(string),
			Provider:        data[StudentServiceProviderField].(string),
			UpdatedAt:       parseTime(data[StudentServiceUpdatedAtField]),
		}
		if err := h.readModel.UpdateStudentService(ctx, projection); err != nil {
			return err
		}

	case StudentServiceDeleted:
		projection := StudentServiceDeletedProjection{
			Position:  resolved.Position,
			ServiceID: serviceID,
			DeletedAt: parseTime(data[StudentServiceDeletedAtField]),
		}
		if err := h.readModel.DeleteStudentService(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, period.Channel(periodID).. etc)
	_ = h.publisher.Publish(ctx, Channel(serviceID), "student service read model update")
	_ = h.publisher.Publish(ctx, se.Channel(studentID), "student service read model update")
	return nil
}
