package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type UpdateStudentCommand struct {
	StudentState
	Metadata CommandMetadata
}

type UpdateStudentResult struct {
	EventID string
	Skipped bool
}

func UpdateStudentCommandHandler(
	ctx context.Context,
	command UpdateStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateStudentResult,
	error,
) {
	model, err := loadUpdateStudentContext(ctx, retriever, command.ID)
	if err != nil {
		return UpdateStudentResult{}, err
	}
	if !model.isActive() {
		return UpdateStudentResult{}, eventstore.ErrStudentNotActive
	}

	// TODO add skip logic

	eventID := uuidv7.NewString()
	event := NewStudentUpdatedEvent(
		eventID,
		command.StudentState,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateStudentResult{}, err
	}
	return UpdateStudentResult{EventID: eventID}, nil
}

type updateStudentContext struct {
	created  bool
	archived bool
	deleted  bool
	StudentState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadUpdateStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID string,
) (*updateStudentContext,
	error,
) {
	query := StreamQuery(studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateStudentContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateStudentContext) isActive() bool {
	if !m.created || m.archived || m.deleted {
		return false
	}
	return true
}

func (m *updateStudentContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EventStudentCreated:
		var event StudentCreatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("student update handle create unmarshal", "err", err)
			return
		}
		m.created = true
		m.StudentState = event.StudentState
	case EventStudentUpdated:
		var event StudentUpdatedEvent
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("student update handle update unmarshal", "err", err)
			return
		}
		m.StudentState = event.StudentState
	case EventStudentArchived:
		m.archived = true
	case EventStudentDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
