package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
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
	model, err := loadDeleteServiceContext(ctx, retriever, cmd.ServiceID, cmd.IEPID, cmd.StudentID)
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
	exists   bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	serviceID,
	iepID,
	studentID string,
) (
	*deleteServiceContext,
	error,
) {
	query := streamQuery(serviceID, studentID)
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
	if !m.exists || m.deleted {
		return false
	}
	return true
}

func (m *deleteServiceContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventServiceAddedToIEP:
		m.exists = true
		m.deleted = false
	case EventServiceDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
