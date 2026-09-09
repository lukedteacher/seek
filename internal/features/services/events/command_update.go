package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"seek/internal/eventstore"
	iepEvents "seek/internal/features/ieps/events"
	"seek/internal/features/services/models"
	"seek/pkg/uuidv7"
)

type UpdateServiceCommand struct {
	Service  models.Service
	Metadata CommandMetadata
}

type UpdateServiceResult struct {
	EventID string
	Skipped bool
}

func UpdateServiceCommandHandler(
	ctx context.Context,
	cmd UpdateServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateServiceResult,
	error,
) {
	model, err := loadUpdateServiceContext(
		ctx,
		retriever,
		cmd.Service.ID,
		cmd.Service.IEPID,
	)
	if err != nil {
		return UpdateServiceResult{}, err
	}
	if !model.isServiceActive() {
		return UpdateServiceResult{}, eventstore.ErrServiceNotActive
	}
	if !model.iep.isActive() {
		return UpdateServiceResult{}, eventstore.ErrIEPNotActive
	}
	// TODO reimpliment this
	// if model.isSame(cmd) {
	// 	return UpdateServiceResult{Skipped: true}, nil
	// }

	eventID := uuidv7.NewString()
	event := NewServiceUpdatedEvent(
		eventID,
		cmd,
		model.query,
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateServiceResult{}, err
	}
	return UpdateServiceResult{EventID: eventID}, nil
}

type updateServiceContext struct {
	serviceExists   bool
	serviceArchived bool
	serviceDeleted  bool
	service         models.Service
	iep             IEPState
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadUpdateServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	serviceID,
	iepID string,
) (
	*updateServiceContext,
	error,
) {
	query := serviceStreamQuery(serviceID, iepID)
	events, err := retriever.GetEvents(
		ctx,
		eventstore.NoEventPosition,
		100,
		eventstore.Forward,
		query,
	)
	if err != nil {
		return nil, err
	}

	model := &updateServiceContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateServiceContext) isServiceActive() bool {
	if m.serviceExists || !m.serviceArchived || !m.serviceDeleted {
		return true
	}
	return false
}

func (m *updateServiceContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case iepEvents.EventIEPAddedToStudent:
		m.iep.created = true
	case iepEvents.EventIEPArchived:
		m.iep.archived = true
	case iepEvents.EventIEPDeleted:
		m.iep.deleted = true
	case EventServiceAddedToIEP:
		m.serviceExists = true
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("service update handle add unmarshal", "err", err)
			return
		}
		m.service = NewModelFromFlat(flat)
	case EventServiceUpdated:
		var flat ServiceFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("service update handle update unmarshal", "err", err)
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
