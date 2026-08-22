package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type UpdateEducatorCommand struct {
	EducatorState
	Metadata CommandMetadata
}

type UpdateEducatorResult struct {
	EventID  string
	Educator EducatorState
	Skipped  bool
}

func UpdateEducatorCommandHandler(
	ctx context.Context,
	cmd UpdateEducatorCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateEducatorResult,
	error,
) {
	model, err := loadUpdateEducatorContext(ctx, retriever, cmd.ID)
	if err != nil {
		return UpdateEducatorResult{}, err
	}
	if !model.isActive() {
		return UpdateEducatorResult{}, eventstore.ErrEducatorNotActive
	}
	// TODO add skip logic

	// build event
	eventData := EducatorUpdatedEvent{
		EventID: uuidv7.NewString(),
		EducatorState: EducatorState{
			ID:         cmd.ID,
			GivenName:  cmd.GivenName,
			ChosenName: cmd.ChosenName,
			FamilyName: cmd.FamilyName,
			Email:      cmd.Email,
			Username:   deriveUsername(cmd.Email),
			Roles:      cmd.Roles,
			UpdatedAt:  time.Now(),
		},
		Scope: educatorScope(cmd.ID),
	}

	// wrap in domain event
	event := eventstore.DomainEvent{
		EventID:   eventData.EventID,
		EventType: EventEducatorUpdated,
		Data:      eventstore.MustData(eventData),
		Metadata:  metadataWithQuery(cmd.Metadata, model.query),
	}

	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		model.events,
		model.query,
	); err != nil {
		return UpdateEducatorResult{}, err
	}
	return UpdateEducatorResult{EventID: eventData.EventID, Educator: eventData.EducatorState}, nil
}

type updateEducatorContext struct {
	created  bool
	archived bool
	deleted  bool
	EducatorState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadUpdateEducatorContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	educatorID string,
) (
	*updateEducatorContext,
	error,
) {
	query := streamQuery(educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &updateEducatorContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateEducatorContext) isActive() bool {
	if !m.created || m.archived || m.deleted {
		return false
	}
	return true
}

func (m *updateEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EventEducatorCreated:
		var event EducatorCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("educator update handle create unmarshal", "err", err)
			return
		}
		m.created = true
		m.EducatorState = event.EducatorState
	case EventEducatorUpdated:
		var event EducatorUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("educator update handle update unmarshal", "err", err)
			return
		}
		m.EducatorState = event.EducatorState
	case EventEducatorArchived:
		m.archived = true
	case EventEducatorDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
