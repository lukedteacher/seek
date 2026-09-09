package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type DeleteIEPCommand struct {
	IEPID     string
	StudentID string
	Metadata  CommandMetadata
}

type DeleteIEPResult struct {
	EventID string
}

func DeleteIEPCommandHandler(
	ctx context.Context,
	cmd DeleteIEPCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	DeleteIEPResult,
	error,
) {
	model, err := loadDeleteIEPContext(ctx, retriever, cmd.IEPID, cmd.StudentID)
	if err != nil {
		return DeleteIEPResult{}, err
	}
	if !model.isActive() {
		return DeleteIEPResult{}, eventstore.ErrIEPNotActive
	}

	eventID := uuidv7.NewString()
	event := NewIEPDeletedEvent(
		eventID,
		cmd.IEPID,
		cmd.StudentID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return DeleteIEPResult{}, err
	}
	return DeleteIEPResult{EventID: eventID}, nil
}

type deleteIEPContext struct {
	exists   bool
	archived bool
	deleted  bool
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadDeleteIEPContext(ctx context.Context, retriever eventstore.Retriever, iepID, studentID string) (*deleteIEPContext, error) {
	query := StreamQuery(iepID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &deleteIEPContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *deleteIEPContext) isActive() bool {
	if !m.exists || m.deleted {
		return false
	}
	return true
}

func (m *deleteIEPContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case EventIEPAddedToStudent:
		m.exists = true
	case EventIEPArchived:
		m.archived = true
	case EventIEPDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
