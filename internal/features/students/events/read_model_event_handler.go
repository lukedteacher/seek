package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/students/models"
)

const StudentReadModelEventHandlerName = "student_read_model_event_handler"

type StudentReadModelReader interface {
	GetByID(ctx context.Context, studentID string) (*models.Student, error)
	GetByUsername(ctx context.Context, username string) (*models.Student, error)
	List(ctx context.Context, opts ...ListOption) ([]models.Student, error)
	ListByServiceType(ctx context.Context, s string) ([]models.Student, error)
}

type StudentReadModelWriter interface {
	Create(ctx context.Context, event StudentCreatedProjection) error
	Update(ctx context.Context, event StudentUpdatedProjection) error
	Archive(ctx context.Context, event StudentArchivedProjection) error
	Delete(ctx context.Context, event StudentDeletedProjection) error
}

type StudentState struct {
	ID         string    `json:"id"`
	MARSSID    string    `json:"marss_id"`
	Birthdate  string    `json:"birthdate"`
	GivenName  string    `json:"given_name"`
	ChosenName string    `json:"chosen_name"`
	FamilyName string    `json:"family_name"`
	Pronouns   string    `json:"pronouns"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	Grade      int       `json:"grade"`
	HomeroomID string    `json:"homeroom_id"`
	PlanType   int       `json:"plan_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type StudentCreatedProjection struct {
	Position eventstore.Position
	StudentState
}

type StudentUpdatedProjection struct {
	Position eventstore.Position
	StudentState
}

type StudentArchivedProjection struct {
	Position   eventstore.Position
	StudentID  string
	ArchivedAt time.Time
}

type StudentDeletedProjection struct {
	Position  eventstore.Position
	StudentID string
	DeletedAt time.Time
}

type StudentReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel StudentReadModelWriter
	publisher eventstore.Publisher
}

func NewStudentReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel StudentReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*StudentReadModelEventHandler,
	error,
) {
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
	eventTypes := []eventType{
		EventStudentCreated,
		EventStudentUpdated,
		EventStudentArchived,
		EventStudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *StudentReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	rawData := resolved.Event.RawData
	scope := eventstore.Scope(data)
	studentID, _ := scope[FieldStudentID].(string)
	switch resolved.Event.EventType {
	case EventStudentCreated:
		var event StudentCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("student read model handle create unmarshal", "err", err)
			return err
		}
		if err := h.readModel.Create(ctx, StudentCreatedProjection{
			Position:     resolved.Position,
			StudentState: event.StudentState,
		}); err != nil {
			return err
		}
	case EventStudentUpdated:
		var event StudentUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("student read model handle update unmarshal", "err", err)
			return err
		}
		if err := h.readModel.Update(ctx, StudentUpdatedProjection{
			Position:     resolved.Position,
			StudentState: event.StudentState,
		}); err != nil {
			return err
		}
	case EventStudentArchived:
		if err := h.readModel.Archive(ctx, StudentArchivedProjection{
			Position:   resolved.Position,
			StudentID:  studentID,
			ArchivedAt: parseTime(data[FieldStudentArchivedAt]),
		}); err != nil {
			return err
		}
	case EventStudentDeleted:
		if err := h.readModel.Delete(ctx, StudentDeletedProjection{
			Position:  resolved.Position,
			StudentID: studentID,
			DeletedAt: parseTime(data[FieldStudentDeletedAt]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled student read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, student.Channel(studentID).. etc)
	return h.publisher.Publish(
		ctx,
		Channel(studentID),
		map[string]string{"studentID": studentID, "type": resolved.Event.EventType.String()},
	)
}

func stringPtr(value string) *string {
	return &value
}
