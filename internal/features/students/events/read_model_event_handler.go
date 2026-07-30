package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/students/models"
)

const StudentReadModelEventHandlerName = "student_read_model_event_handler"

type StudentReadModelReader interface {
	Get(ctx context.Context, studentID string) (*models.Student, error)
	List(ctx context.Context) ([]models.Student, error)
	ListStudentsByIEPServiceType(ctx context.Context, s string) ([]models.Student, error)
}

type StudentReadModelWriter interface {
	CreateStudent(ctx context.Context, event StudentCreatedProjection) error
	UpdateStudent(ctx context.Context, event StudentUpdatedProjection) error
	DeleteStudent(ctx context.Context, event StudentDeletedProjection) error
}

type StudentCreatedProjection struct {
	Position    eventstore.Position
	StudentID   string
	GivenName   string
	ChosenName  string
	FamilyName  string
	Grade       int
	Homeroom    string
	CaseManager string
	CreatedAt   time.Time
}

type StudentUpdatedProjection struct {
	Position    eventstore.Position
	StudentID   string
	GivenName   string
	ChosenName  string
	FamilyName  string
	Grade       int
	Homeroom    string
	CaseManager string
	UpdatedAt   time.Time
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
	studentID, _ := scope[FieldStudentID].(string)
	switch resolved.Event.EventType {
	case StudentCreated:
		givenName, _ := data[FieldStudentGivenName].(string)
		chosenName, _ := data[FieldStudentChosenName].(string)
		familyName, _ := data[FieldStudentFamilyName].(string)
		grade := int(data[FieldStudentGrade].(float64))
		homeroom, _ := data[FieldStudentHomeroom].(string)
		caseManager, _ := data[FieldStudentCaseManager].(string)
		if err := h.readModel.CreateStudent(ctx, StudentCreatedProjection{
			Position:    resolved.Position,
			StudentID:   studentID,
			GivenName:   givenName,
			ChosenName:  chosenName,
			FamilyName:  familyName,
			Grade:       grade,
			Homeroom:    homeroom,
			CaseManager: caseManager,
			CreatedAt:   parseTime(data[FieldStudentCreatedAt]),
		}); err != nil {
			return err
		}
	case StudentUpdated:
		givenName, _ := data[FieldStudentGivenName].(string)
		chosenName, _ := data[FieldStudentChosenName].(string)
		familyName, _ := data[FieldStudentFamilyName].(string)
		grade := int(data[FieldStudentGrade].(float64))
		homeroom, _ := data[FieldStudentHomeroom].(string)
		caseManager, _ := data[FieldStudentCaseManager].(string)
		if err := h.readModel.UpdateStudent(ctx, StudentUpdatedProjection{
			Position:    resolved.Position,
			StudentID:   studentID,
			GivenName:   givenName,
			ChosenName:  chosenName,
			FamilyName:  familyName,
			Grade:       grade,
			Homeroom:    homeroom,
			CaseManager: caseManager,
			UpdatedAt:   parseTime(data[FieldStudentUpdatedAt]),
		}); err != nil {
			return err
		}
	case StudentDeleted:
		if err := h.readModel.DeleteStudent(ctx, StudentDeletedProjection{
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
		map[string]string{"periodID": studentID},
	)
}

func stringPtr(value string) *string {
	return &value
}
