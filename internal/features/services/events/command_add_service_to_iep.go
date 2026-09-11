package events

import (
	"context"

	"seek/internal/eventstore"
	iepEvents "seek/internal/features/ieps/events"
	"seek/internal/features/services/models"
	"seek/pkg/uuidv7"
)

type AddServiceToIEPCommand struct {
	Service  models.Service
	Metadata CommandMetadata
}

type AddServiceToIEPResult struct {
	EventID string
	Skipped bool
}

func AddServiceToIEPCommandHandler(
	ctx context.Context,
	cmd AddServiceToIEPCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddServiceToIEPResult,
	error,
) {
	model, err := loadAddServiceToIEPContext(
		ctx,
		retriever,
		cmd.Service.IEPID,
	)
	if err != nil {
		return &AddServiceToIEPResult{}, err
	}
	if !model.iep.isActive() {
		return &AddServiceToIEPResult{}, eventstore.ErrIEPNotActive
	}
	eventID := uuidv7.NewString()
	cmd.Service.ID = eventID
	event := NewServiceAddedToStudentEvent(
		cmd,
		model.query,
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		nil,
		model.query,
	); err != nil {
		return &AddServiceToIEPResult{}, err
	}
	return &AddServiceToIEPResult{EventID: eventID, Skipped: false}, nil
}

type addServiceToIEPContext struct {
	iep      IEPState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadAddServiceToIEPContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	iepID string,
) (
	*addServiceToIEPContext,
	error,
) {
	query := iepStreamQuery(iepID)
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
	model := &addServiceToIEPContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	// creates a model of the relevant context from past events
	for i := range events {
		model.handle(events[i])
	}
	return model, nil
}

func (m *addServiceToIEPContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case iepEvents.EventIEPAddedToStudent:
		m.iep.created = true
	case iepEvents.EventIEPArchived:
		m.iep.archived = true
	case iepEvents.EventIEPDeleted:
		m.iep.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
