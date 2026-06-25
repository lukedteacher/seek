package teacher

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
)

const TeacherReadModelEventHandlerName = "teacher_read_model_event_handler"

type TeacherReadModelReader interface {
	Get(ctx context.Context, teacherID string) (*models.Teacher, error)
	List(ctx context.Context) ([]models.Teacher, error)
}

type TeacherReadModelWriter interface {
	InsertCreatedTeacher(ctx context.Context, event TeacherCreatedProjection) error
	UpdateTeacher(ctx context.Context, event TeacherUpdatedProjection) error
	DeleteTeacher(ctx context.Context, event TeacherDeletedProjection) error
}

type TeacherCreatedProjection struct {
	Position   eventstore.Position
	Id         string
	FirstName  string
	ChosenName *string
	LastName   string
	CreatedAt  time.Time
}

type TeacherUpdatedProjection struct {
	Position   eventstore.Position
	Id         string
	FirstName  string
	ChosenName *string
	LastName   string
	UpdatedAt  time.Time
}

type TeacherDeletedProjection struct {
	Position  eventstore.Position
	Id        string
	DeletedAt time.Time
}

type TeacherReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel TeacherReadModelWriter
	publisher eventstore.Publisher
}

func NewTeacherReadModelEventHandler(subscriber eventstore.Subscriber, checkpointer eventstore.Checkpointer, readModel TeacherReadModelWriter, publisher eventstore.Publisher, logger *slog.Logger) (*TeacherReadModelEventHandler, error) {
	handler := &TeacherReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            TeacherReadModelEventHandlerName,
		Query:           TeacherReadModelEventHandlerQuery(),
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

func (h *TeacherReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *TeacherReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func TeacherReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		TeacherCreated,
		TeacherUpdated,
		TeacherDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *TeacherReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	id, _ := scope["id"].(string)
	switch resolved.Event.EventType {
	case TeacherCreated:
		firstName, _ := data["firstName"].(string)
		chosenName, _ := data["chosenName"].(*string)
		lastName, _ := data["lastName"].(string)
		if err := h.readModel.InsertCreatedTeacher(ctx, TeacherCreatedProjection{
			Position:   resolved.Position,
			Id:         id,
			FirstName:  firstName,
			ChosenName: chosenName,
			LastName:   lastName,
			CreatedAt:  parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case TeacherUpdated:
		firstName, _ := data["firstName"].(string)
		chosenName, _ := data["chosenName"].(*string)
		lastName, _ := data["lastName"].(string)
		if err := h.readModel.UpdateTeacher(ctx, TeacherUpdatedProjection{
			Position:   resolved.Position,
			Id:         id,
			FirstName:  firstName,
			ChosenName: chosenName,
			LastName:   lastName,
			UpdatedAt:  parseTime(data["updatedAt"]),
		}); err != nil {
			return err
		}
	case TeacherDeleted:
		if err := h.readModel.DeleteTeacher(ctx, TeacherDeletedProjection{
			Position:  resolved.Position,
			Id:        id,
			DeletedAt: parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled teacher read model event type %q", resolved.Event.EventType)
	}
	// TODO not sure if this is the fix
	return nil
}
