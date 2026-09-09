package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/ieps/models"
	se "seek/internal/features/students/events"
)

const IEPReadModelEventHandlerName = "iep_read_model_event_handler"

type IEPReadModelReader interface {
	Get(ctx context.Context, iepID string) (*models.IEP, error)
	List(ctx context.Context) ([]models.IEP, error)
	ListForStudent(ctx context.Context, studentID string) ([]models.IEP, error)
}

type IEPReadModelWriter interface {
	AddIEPToStudent(ctx context.Context, event IEPAddedToStudentProjection) error
	UpdateIEP(ctx context.Context, event IEPUpdatedProjection) error
	ArchiveIEP(ctx context.Context, event IEPArchivedProjection) error
	DeleteIEP(ctx context.Context, event IEPDeletedProjection) error
}

type IEPAddedToStudentProjection struct {
	Position eventstore.Position
	models.IEP
}

type IEPUpdatedProjection struct {
	Position eventstore.Position
	models.IEP
}

type IEPArchivedProjection struct {
	Position   eventstore.Position
	IEPID      string
	ArchivedAt time.Time
}

type IEPDeletedProjection struct {
	Position eventstore.Position
	IEPID    string
}

type IEPReadModelEventHandler struct {
	global    *eventstore.GlobalEventHandler
	readModel IEPReadModelWriter
	publisher eventstore.Publisher
}

func NewIEPReadModelEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	readModel IEPReadModelWriter,
	publisher eventstore.Publisher,
	logger *slog.Logger,
) (
	*IEPReadModelEventHandler,
	error,
) {
	handler := &IEPReadModelEventHandler{readModel: readModel, publisher: publisher}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            IEPReadModelEventHandlerName,
		Query:           IEPReadModelEventHandlerQuery(),
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

func (h *IEPReadModelEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *IEPReadModelEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func IEPReadModelEventHandlerQuery() eventstore.Query {
	eventTypes := []eventType{
		EventIEPAddedToStudent,
		EventIEPRemovedFromStudent,
		EventIEPUpdated,
		EventIEPArchived,
		EventIEPDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{{Key: eventTypeKey, Value: eventType.String()}},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func (h *IEPReadModelEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	data := resolved.Event.Data
	rawData := resolved.Event.RawData
	scope := eventstore.Scope(data)
	iepID, _ := scope[FieldIEPID].(string)
	studentID, _ := scope[FieldIEPStudentID].(string)
	switch resolved.Event.EventType {
	case EventIEPAddedToStudent:
		var flat IEPFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			return err
		}
		model := NewModelFromFlat(flat)
		projection := IEPAddedToStudentProjection{
			Position: resolved.Position,
			IEP:      model,
		}
		if err := h.readModel.AddIEPToStudent(ctx, projection); err != nil {
			return err
		}
	case EventIEPUpdated:
		var flat IEPFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			return err
		}
		model := NewModelFromFlat(flat)
		projection := IEPUpdatedProjection{
			Position: resolved.Position,
			IEP:      model,
		}
		if err := h.readModel.UpdateIEP(ctx, projection); err != nil {
			return err
		}
	case EventIEPArchived:
		var event IEPArchivedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep read model handle archive unmarshal", "err", err)
		}
		projection := IEPArchivedProjection{
			Position: resolved.Position,
			IEPID:    iepID,
		}
		if err := h.readModel.ArchiveIEP(ctx, projection); err != nil {
			return err
		}
	case EventIEPDeleted:
		var event IEPDeletedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep read model handle delete unmarshal", "err", err)
		}
		projection := IEPDeletedProjection{
			Position: resolved.Position,
			IEPID:    iepID,
		}
		if err := h.readModel.DeleteIEP(ctx, projection); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled period read model event type %q", resolved.Event.EventType)
	}
	// so the SSE stream will update
	// s.Subscriber.Subscribe(ctx, iep.Channel(iepID) ...etc.)
	_ = h.publisher.Publish(ctx, Channel(iepID), "iep read model update")
	_ = h.publisher.Publish(ctx, se.Channel(studentID), "student read model update")
	return nil
}

func unflatten(flat map[string]interface{}) (map[string]interface{}, error) {
	unflat := map[string]interface{}{}

	for key, value := range flat {
		keyParts := strings.Split(key, ".")

		// walk the keys until we get to a leaf node.
		m := unflat
		for i, k := range keyParts[:len(keyParts)-1] {
			v, exists := m[k]
			if !exists {
				newMap := map[string]interface{}{}
				m[k] = newMap
				m = newMap
				continue
			}

			innerMap, ok := v.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("key=%v is not an object", strings.Join(keyParts[0:i+1], "."))
			}
			m = innerMap
		}

		leafKey := keyParts[len(keyParts)-1]
		if _, exists := m[leafKey]; exists {
			return nil, fmt.Errorf("key=%v already exists", key)
		}
		m[keyParts[len(keyParts)-1]] = value
	}

	return unflat, nil
}
