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
	EventID string
	Skipped bool
}

func UpdateEducatorCommandHandler(
	ctx context.Context,
	command UpdateEducatorCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateEducatorResult,
	error,
) {
	model, err := loadUpdateEducatorContext(ctx, retriever, command.ID)
	if err != nil {
		return UpdateEducatorResult{}, err
	}
	if !model.isActive() {
		return UpdateEducatorResult{}, eventstore.ErrEducatorNotActive
	}
	// TODO add skip logic

	// create the event id since we'll be needing it shortly
	eventID := uuidv7.NewString()

	// build event data struct directly
	eventData := EducatorUpdatedEvent{
		EventID: eventID,
		EducatorState: EducatorState{
			GivenName:  command.GivenName,
			ChosenName: command.ChosenName,
			FamilyName: command.FamilyName,
			Email:      command.Email,
			Username:   deriveUsername(command.Email),
			Roles:      command.Roles,
			UpdatedAt:  time.Now(),
		},
		Scope: educatorScope(model.id),
	}

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventEducatorUpdated,
		Data:      eventstore.MustData(eventData),
		Metadata:  metadataWithQuery(command.Metadata, model.query),
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
	return UpdateEducatorResult{EventID: eventID}, nil
}

type updateEducatorContext struct {
	created  bool
	archived bool
	deleted  bool
	id       string
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
		id:       educatorID,
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
