package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/homerooms/models"
	"seek/pkg/uuidv7"
)

type UpdateHomeroomCommand struct {
	Homeroom models.Homeroom
	Metadata CommandMetadata
}

type UpdateHomeroomResult struct {
	HomeroomUpdatedID string
	Skipped           bool
}

func UpdateHomeroomCommandHandler(
	ctx context.Context,
	cmd UpdateHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateHomeroomResult,
	error,
) {
	model, err := loadUpdateHomeroomContext(ctx, retriever, cmd.Homeroom.ID)
	if err != nil {
		return UpdateHomeroomResult{}, err
	}
	if err := model.isActive(); err != nil {
		return UpdateHomeroomResult{}, err
	}
	if model.isSame(cmd) {
		return UpdateHomeroomResult{Skipped: true}, nil
	}
	eventID := uuidv7.NewString()
	event := NewHomeroomUpdatedEvent(
		cmd.Homeroom,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateHomeroomResult{}, err
	}
	return UpdateHomeroomResult{HomeroomUpdatedID: eventID}, nil
}

type updateHomeroomContext struct {
	exists   bool
	archived bool
	deleted  bool
	homeroom models.Homeroom
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
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
	return m.homeroom.Title == cmd.Homeroom.Title &&
		m.homeroom.GradesBitmask == cmd.Homeroom.GradesBitmask &&
		m.homeroom.LocationID == cmd.Homeroom.LocationID &&
		m.homeroom.Image == cmd.Homeroom.Image
}

func (m *updateHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		var event HomeroomCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("homeroom update handle create unmarshal", "err", err)
			return
		}
		m.exists = true
		m.homeroom = models.Homeroom{
			ID:            event.Scope.ID,
			Title:         event.Title,
			GradesBitmask: sharedmodels.GradesBitmask(event.GradesBitmask),
			LocationID:    event.LocationID,
			Image:         event.LocationID,
		}
	case EventHomeroomUpdated:
		var event HomeroomUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("homeroom update handle update unmarshal", "err", err)
			return
		}
		m.homeroom = models.Homeroom{
			ID:            event.Scope.ID,
			Title:         event.Title,
			GradesBitmask: sharedmodels.GradesBitmask(event.GradesBitmask),
			LocationID:    event.LocationID,
			Image:         event.LocationID,
		}
	case EventHomeroomArchived:
		m.archived = true
	case EventHomeroomDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
