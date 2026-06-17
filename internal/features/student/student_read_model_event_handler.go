package student

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/views"
)

const StudentReadModelEventHandlerName = "student_read_model_event_handler"

type StudentReadModelReader interface {
	List(ctx context.Context, userRegisteredID string) ([]views.Student, error)
}

type StudentReadModelWriter interface {
	InsertCreatedStudent(ctx context.Context, event StudentCreatedProjection) error
	RenameStudent(ctx context.Context, event StudentRenamedProjection) error
	DeleteStudent(ctx context.Context, event StudentDeletedProjection) error
}

type StudentCreatedProjection struct {
	Position         eventstore.Position
	StudentID           string
	UserRegisteredID string
	FirstName            string
	CreatedAt        time.Time
}

type StudentRenamedProjection struct {
	Position  eventstore.Position
	StudentID    string
	FirstName     string
	RenamedAt time.Time
}

type StudentCompletedProjection struct {
	Position    eventstore.Position
	StudentID      string
	CompletedAt time.Time
}

type StudentReopenedProjection struct {
	Position   eventstore.Position
	StudentID     string
	ReopenedAt time.Time
}

type StudentDeletedProjection struct {
	Position  eventstore.Position
	StudentID    string
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
	criteria := make([]eventstore.Criterion, 0, 5)
	for _, eventType := range []string{StudentCreated, StudentRenamed, StudentCompleted, StudentReopened, StudentDeleted} {
		criteria = append(criteria, eventstore.Criterion{Tags: []eventstore.Tag{{Key: "eventType", Value: eventType}}})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *StudentReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	studentID, _ := scope["studentId"].(string)
	userRegisteredID, _ := scope["userRegisteredId"].(string)

	switch resolved.Event.EventType {
	case StudentCreated:
		firstName, _ := data["first_name"].(string)
		if err := h.readModel.InsertCreatedStudent(ctx, StudentCreatedProjection{
			Position:         resolved.Position,
			StudentID:           studentID,
			UserRegisteredID: userRegisteredID,
			FirstName:            firstName,
			CreatedAt:        parseTime(data["createdAt"]),
		}); err != nil {
			return err
		}
	case StudentRenamed:
		firstName, _ := data["first_name"].(string)
		if err := h.readModel.RenameStudent(ctx, StudentRenamedProjection{
			Position:  resolved.Position,
			StudentID:    studentID,
			FirstName:     firstName,
			RenamedAt: parseTime(data["renamedAt"]),
		}); err != nil {
			return err
		}
	case StudentDeleted:
		if err := h.readModel.DeleteStudent(ctx, StudentDeletedProjection{
			Position:  resolved.Position,
			StudentID:    studentID,
			DeletedAt: parseTime(data["deletedAt"]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled student read model event type %q", resolved.Event.EventType)
	}

	if userRegisteredID == "" {
		return nil
	}
	return h.publisher.Publish(ctx, Channel(userRegisteredID), map[string]string{"userRegisteredId": userRegisteredID})
}
