package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	iepEvents "seek/internal/features/ieps/events"
	"seek/internal/features/services/models"
	"seek/pkg/uuidv7"
)

type DeleteServiceCommand struct {
	ServiceID string
	IEPID     string
	StudentID string
	Metadata  CommandMetadata
}

type DeleteServiceResult struct {
	EventID string
}

func DeleteServiceCommandHandler(
	ctx context.Context,
	cmd DeleteServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteServiceResult,
	error,
) {
	model, err := loadDeleteServiceContext(
		ctx,
		retriever,
		cmd.ServiceID,
		cmd.IEPID,
	)
	if err != nil {
		return DeleteServiceResult{}, err
	}
	if !model.isActive() {
		return DeleteServiceResult{}, eventstore.ErrServiceNotActive
	}

	eventID := uuidv7.NewString()
	event := NewServiceDeletedEvent(
		eventID,
		cmd.ServiceID,
		cmd.IEPID,
		cmd.StudentID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteServiceResult{}, err
	}
	return DeleteServiceResult{EventID: eventID}, nil
}

type deleteServiceContext struct {
	serviceCreated  bool
	serviceArchived bool
	serviceDeleted  bool
	service         models.Service
	iep             IEPState
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadDeleteServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	serviceID,
	iepID string,
) (
	*deleteServiceContext,
	error,
) {
	query := serviceStreamQuery(serviceID, iepID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &deleteServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteServiceContext) isActive() bool {
	if !m.serviceCreated || m.serviceArchived || m.serviceDeleted {
		return false
	}
	return true
}

func (m *deleteServiceContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case iepEvents.EventIEPAddedToStudent:
		m.iep.created = true
	case iepEvents.EventIEPArchived:
		m.iep.archived = true
	case iepEvents.EventIEPDeleted:
		m.iep.deleted = true
	case EventServiceAddedToIEP:
		m.serviceCreated = true
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("service delete handle add unmarshal", "err", err)
			return
		}
		m.service = NewModelFromFlat(flat)
	case EventServiceUpdated:
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("service delete handle update unmarshal", "err", err)
			return
		}
		m.service = NewModelFromFlat(flat)
	case EventServiceArchived:
		m.serviceArchived = true
	case EventServiceDeleted:
		m.serviceDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
