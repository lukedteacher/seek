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
	EducatorID string
	GivenName  string
	ChosenName string
	FamilyName string
	Email      string
	Roles       []string
	Metadata   CommandMetadata
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
	model, err := loadUpdateEducatorContext(ctx, retriever, command.EducatorID)
	if err != nil {
		return UpdateEducatorResult{}, err
	}
	if err := model.isActive(); err != nil {
		return UpdateEducatorResult{}, err
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
			Roles:       command.Roles,
		},
		UpdatedAt: time.Now(),
		Scope:     educatorScope(model.id),
	}

	// wrap data in a domain event
	event := eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EducatorUpdated,
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

func (m *updateEducatorContext) isActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrNotActive
	}
	return nil
}

func (m *updateEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EducatorCreated:
		var event EducatorCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("educator update handle create unmarshal", "err", err)
			return
		}
		m.created = true
		m.archived = false
		m.deleted = false
		m.EducatorState = event.EducatorState
	case EducatorUpdated:
		var event EducatorUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("educator update handle update unmarshal", "err", err)
			return
		}
		m.EducatorState = event.EducatorState
	case EducatorArchived:
		m.archived = true
	case EducatorDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
