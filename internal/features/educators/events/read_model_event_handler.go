package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/educators/models"
)

const EducatorReadModelEventHandlerName = "educator_read_model_event_handler"

type EducatorReadModelReader interface {
	Get(ctx context.Context, educatorID string) (*models.Educator, error)
	List(ctx context.Context) ([]models.Educator, error)
}

type EducatorReadModelWriter interface {
	CreateEducator(ctx context.Context, event EducatorCreatedProjection) error
	UpdateEducator(ctx context.Context, event EducatorUpdatedProjection) error
	ArchiveEducator(ctx context.Context, event EducatorArchivedProjection) error
	DeleteEducator(ctx context.Context, event EducatorDeletedProjection) error
}

type EducatorCreatedProjection struct {
	Position   eventstore.Position
	ID         string
	GivenName  string
	ChosenName string
	FamilyName string
	Role       string
	Email      string
	CreatedAt  time.Time
}

type EducatorUpdatedProjection struct {
	Position   eventstore.Position
	ID         string
	GivenName  string
	ChosenName string
	FamilyName string
	Role       string
	Email      string
	UpdatedAt  time.Time
}

type EducatorArchivedProjection struct {
	Position   eventstore.Position
	ID         string
	ArchivedAt time.Time
}

type EducatorDeletedProjection struct {
	Position  eventstore.Position
	ID        string
	DeletedAt time.Time
}

type EducatorReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel EducatorReadModelWriter
	publisher eventstore.Publisher
}

func NewReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel EducatorReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*EducatorReadModelEventHandler,
	error,
) {
	handler := &EducatorReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            EducatorReadModelEventHandlerName,
		Query:           EducatorReadModelEventHandlerQuery(),
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

func (h *EducatorReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *EducatorReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func EducatorReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		EducatorCreated,
		EducatorUpdated,
		EducatorArchived,
		EducatorDeleted,
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

func (h *EducatorReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	scope := eventstore.Scope(data)
	educatorCreatedEventID, _ := scope[EducatorCreatedEventIDField].(string)
	switch resolved.Event.EventType {
	case EducatorCreated:
		givenName, _ := data[EducatorGivenNameField].(string)
		chosenName, _ := data[EducatorChosenNameField].(string)
		familyName, _ := data[EducatorFamilyNameField].(string)
		role, _ := data[EducatorRoleField].(string)
		email, _ := data[EducatorEmailField].(string)
		if err := h.readModel.CreateEducator(ctx, EducatorCreatedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			GivenName:  givenName,
			ChosenName: chosenName,
			FamilyName: familyName,
			Role:       role,
			Email:      email,
			CreatedAt:  parseTime(data[EducatorCreatedAtField]),
		}); err != nil {
			return err
		}
	case EducatorUpdated:
		givenName, _ := data[EducatorGivenNameField].(string)
		chosenName, _ := data[EducatorChosenNameField].(string)
		familyName, _ := data[EducatorFamilyNameField].(string)
		role, _ := data[EducatorRoleField].(string)
		email, _ := data[EducatorEmailField].(string)
		if err := h.readModel.UpdateEducator(ctx, EducatorUpdatedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			GivenName:  givenName,
			ChosenName: chosenName,
			FamilyName: familyName,
			Role:       role,
			Email:      email,
			UpdatedAt:  parseTime(data[EducatorUpdatedAtField]),
		}); err != nil {
			return err
		}
	case EducatorArchived:
		if err := h.readModel.ArchiveEducator(ctx, EducatorArchivedProjection{
			Position:   resolved.Position,
			ID:         educatorCreatedEventID,
			ArchivedAt: parseTime(data[EducatorArchivedAtField]),
		}); err != nil {
			return err
		}
	case EducatorDeleted:
		if err := h.readModel.DeleteEducator(ctx, EducatorDeletedProjection{
			Position:  resolved.Position,
			ID:        educatorCreatedEventID,
			DeletedAt: parseTime(data[EducatorDeletedAtField]),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled educator read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, educator.Channel(educatorID).. etc)
	return h.publisher.Publish(
		ctx,
		Channel(educatorCreatedEventID),
		map[string]string{"educatorID": educatorCreatedEventID}, // TODO figure out why it is this
	)
}

func stringPtr(value string) *string {
	return &value
}
