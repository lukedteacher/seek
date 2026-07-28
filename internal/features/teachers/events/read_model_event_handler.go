package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/teachers/models"
)

const TeacherReadModelEventHandlerName = "teacher_read_model_event_handler"

type TeacherReadModelReader interface {
	Get(ctx context.Context, teacherID string) (*models.Teacher, error)
	List(ctx context.Context) ([]models.Teacher, error)
}

type TeacherReadModelWriter interface {
	CreateTeacher(ctx context.Context, event TeacherCreatedProjection) error
	UpdateTeacher(ctx context.Context, event TeacherUpdatedProjection) error
	DeleteTeacher(ctx context.Context, event TeacherDeletedProjection) error
}

type TeacherCreatedProjection struct {
	Position   eventstore.Position
	TeacherID  string
	GivenName  string
	ChosenName *string
	FamilyName string
	CreatedAt  time.Time
}

type TeacherUpdatedProjection struct {
	Position   eventstore.Position
	TeacherID  string
	GivenName  string
	ChosenName *string
	FamilyName string
	UpdatedAt  time.Time
}

type TeacherDeletedProjection struct {
	Position  eventstore.Position
	TeacherID string
	DeletedAt time.Time
}

type TeacherReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel TeacherReadModelWriter
	publisher eventstore.Publisher
}

func NewTeacherReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel TeacherReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*TeacherReadModelEventHandler,
	error,
) {
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
	teacherID, _ := scope[TeacherIDField].(string)
	switch resolved.Event.EventType {
	case TeacherCreated:
		givenName, _ := data[TeacherGivenNameField].(string)
		chosenName, _ := data[TeacherChosenNameField].(string)
		familyName, _ := data[TeacherFamilyNameField].(string)
		if err := h.readModel.CreateTeacher(ctx, TeacherCreatedProjection{
			Position:   resolved.Position,
			TeacherID:  teacherID,
			GivenName:  givenName,
			ChosenName: stringPtr(chosenName),
			FamilyName: familyName,
			CreatedAt:  parseTime(data[TeacherCreatedAtField]),
		}); err != nil {
			return err
		}
	case TeacherUpdated:
		givenName, _ := data[TeacherGivenNameField].(string)
		chosenName, _ := data[TeacherChosenNameField].(string)
		familyName, _ := data[TeacherFamilyNameField].(string)
		if err := h.readModel.UpdateTeacher(ctx, TeacherUpdatedProjection{
			Position:   resolved.Position,
			TeacherID:  teacherID,
			GivenName:  givenName,
			ChosenName: stringPtr(chosenName),
			FamilyName: familyName,
			UpdatedAt:  parseTime(data[TeacherUpdatedAtField]),
		}); err != nil {
			return err
		}
	case TeacherDeleted:
		if err := h.readModel.DeleteTeacher(ctx, TeacherDeletedProjection{
			Position:  resolved.Position,
			TeacherID: teacherID,
			DeletedAt: parseTime(data[TeacherDeletedAtField]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled teacher read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, teacher.Channel(teacherID).. etc)
	return h.publisher.Publish(
		ctx,
		Channel(teacherID),
		map[string]string{"periodID": teacherID},
	)
}

func stringPtr(value string) *string {
	return &value
}
