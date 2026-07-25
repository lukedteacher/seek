package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type UpdateEducatorCommand struct {
	ID         string
	GivenName  string
	ChosenName string
	FamilyName string
	Email      string
	Role       string
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
	model, err := loadUpdateEducatorContext(ctx, retriever, command.ID)
	if err != nil {
		return UpdateEducatorResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return UpdateEducatorResult{}, err
	}
	// TODO add skip logic

	eventID := uuidv7.NewString()
	event := NewEducatorUpdatedEvent(
		eventID,
		command.ID,
		command.GivenName,
		command.ChosenName,
		command.FamilyName,
		command.Email,
		command.Role,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

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
	exists     bool
	deleted    bool
	givenName  string
	chosenName string
	familyName string
	email      string
	role       string
	position   eventstore.Position
	events     []eventstore.ResolvedEvent
	query      eventstore.Query
}

func loadUpdateEducatorContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	id string,
) (
	*updateEducatorContext,
	error,
) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateEducatorContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateEducatorContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case EducatorCreated:
		m.exists = true
		m.deleted = false
		m.givenName, _ = data[EducatorGivenNameField].(string)
		m.chosenName, _ = data[EducatorChosenNameField].(string)
		m.familyName, _ = data[EducatorFamilyNameField].(string)
		m.email = data[EducatorEmailField].(string)
		m.role = data[EducatorRoleField].(string)
	case EducatorUpdated:
		m.givenName, _ = data[EducatorGivenNameField].(string)
		m.chosenName, _ = data[EducatorChosenNameField].(string)
		m.familyName, _ = data[EducatorFamilyNameField].(string)
		m.email = data[EducatorEmailField].(string)
		m.role = data[EducatorRoleField].(string)
	case EducatorDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
