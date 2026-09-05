package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type UpdateHomeroomCommand struct {
	ID         string
	Title      string
	LocationID string
	Metadata   CommandMetadata
}

type UpdateHomeroomResult struct {
	HomeroomUpdatedID string
	Skipped           bool
}

func UpdateHomeroomCommandHandler(
	ctx context.Context,
	command UpdateHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateHomeroomResult,
	error,
) {
	model, err := loadUpdateHomeroomContext(ctx, retriever, command.ID)
	if err != nil {
		return UpdateHomeroomResult{}, err
	}
	if err := model.isActive(); err != nil {
		return UpdateHomeroomResult{}, err
	}
	if model.isSame(command) {
		return UpdateHomeroomResult{Skipped: true}, nil
	}
	eventID := uuidv7.NewString()
	event := NewHomeroomUpdatedEvent(
		command.ID,
		command.Title,
		command.LocationID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateHomeroomResult{}, err
	}
	return UpdateHomeroomResult{HomeroomUpdatedID: eventID}, nil
}

type updateHomeroomContext struct {
	exists     bool
	archived   bool
	deleted    bool
	title      string
	locationID string
	position   eventstore.Position
	events     []eventstore.ResolvedEvent
	query      eventstore.Query
}

func loadUpdateHomeroomContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updateHomeroomContext, error) {
	query := homeroomStreamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateHomeroomContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateHomeroomContext) isActive() error {
	if !m.exists || m.archived || m.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func (m *updateHomeroomContext) isSame(cmd UpdateHomeroomCommand) bool {
	return m.title == cmd.Title &&
		m.locationID == cmd.LocationID
}

func (m *updateHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.exists = true
		m.title, _ = data[FieldHomeroomTitle].(string)
		m.locationID, _ = data[FieldHomeroomLocationID].(string)
	case EventHomeroomUpdated:
		m.title, _ = data[FieldHomeroomTitle].(string)
		m.locationID, _ = data[FieldHomeroomLocationID].(string)
	case EventHomeroomArchived:
		m.archived = true
	case EventHomeroomDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
