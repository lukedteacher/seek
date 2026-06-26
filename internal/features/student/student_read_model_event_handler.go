package student

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
)

const StudentReadModelEventHandlerName = "student_read_model_event_handler"

type StudentReadModelReader interface {
	Get(ctx context.Context, studentID string) (*models.Student, error)
	List(ctx context.Context) ([]models.Student, error)
}

type StudentReadModelWriter interface {
	InsertCreatedStudent(ctx context.Context, event StudentCreatedProjection) error
	UpdateStudent(ctx context.Context, event StudentUpdatedProjection) error
	DeleteStudent(ctx context.Context, event StudentDeletedProjection) error
}

type StudentCreatedProjection struct {
	Position    eventstore.Position
	Id          string
	FirstName   string
	ChosenName  *string
	LastName    string
	Grade       int64
	Homeroom    string
	CaseManager *string
	CreatedAt   time.Time
}

type StudentUpdatedProjection struct {
	Position    eventstore.Position
	Id          string
	FirstName   string
	ChosenName  *string
	LastName    string
	Grade       int64
	Homeroom    string
	CaseManager *string
	UpdatedAt   time.Time
}

type StudentDeletedProjection struct {
	Position  eventstore.Position
	Id        string
	DeletedAt time.Time
}

type StudentReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel StudentReadModelWriter
	publisher eventstore.Publisher
}

func NewStudentReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel StudentReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*StudentReadModelEventHandler, error) {
	handler := &StudentReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            StudentReadModelEventHandlerName,
		Query:           StudentReadModelEventHandlerQuery(),
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

func (h *StudentReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *StudentReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func StudentReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		StudentCreated,
		StudentUpdated,
		StudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *StudentReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	id, _ := scope["id"].(string)

	switch resolved.Event.EventType {
	case StudentCreated:
		firstName, _ := data["firstName"].(string)
		chosenName, _ := data["chosenName"].(string)
		lastName, _ := data["lastName"].(string)
		grade := int64(data["grade"].(float64))
		homeroom, _ := data["homeroom"].(string)
		caseManager, _ := data["caseManager"].(string)
		if err := h.readModel.InsertCreatedStudent(ctx, StudentCreatedProjection{
			Position:    resolved.Position,
			Id:          id,
			FirstName:   firstName,
			ChosenName:  stringPtr(chosenName),
			LastName:    lastName,
			Grade:       grade,
			Homeroom:    homeroom,
			CaseManager: stringPtr(caseManager),
			CreatedAt:   parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case StudentUpdated:
		firstName, _ := data["firstName"].(string)
		chosenName, _ := data["chosenName"].(string)
		lastName, _ := data["lastName"].(string)
		grade := int64(data["grade"].(float64))
		homeroom, _ := data["homeroom"].(string)
		caseManager, _ := data["caseManager"].(string)
		if err := h.readModel.UpdateStudent(ctx, StudentUpdatedProjection{
			Position:    resolved.Position,
			Id:          id,
			FirstName:   firstName,
			ChosenName:  stringPtr(chosenName),
			LastName:    lastName,
			Grade:       grade,
			Homeroom:    homeroom,
			CaseManager: stringPtr(caseManager),
			UpdatedAt:   parseTime(data["updatedAt"]),
		}); err != nil {
			return err
		}
	case StudentDeleted:
		if err := h.readModel.DeleteStudent(ctx, StudentDeletedProjection{
			Position:  resolved.Position,
			Id:        id,
			DeletedAt: parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled student read model event type %q", resolved.Event.EventType)
	}
	// TODO not sure if this is the fix
	return nil
}

func stringPtr(value string) *string {
	return &value
}
