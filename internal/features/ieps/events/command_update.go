package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type UpdateIEPCommand struct {
	IEP      IEPState
	Metadata CommandMetadata
}

type UpdateIEPResult struct {
	EventID string
	Skipped bool
}

func UpdateIEPCommandHandler(
	ctx context.Context,
	command UpdateIEPCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateIEPResult,
	error,
) {
	model, err := loadUpdateStudentIEPContext(ctx, retriever, command.IEP.ID, command.IEP.StudentID)
	if err != nil {
		return UpdateIEPResult{}, err
	}
	if err := model.isIEPActive(); err != nil {
		return UpdateIEPResult{}, err
	}
	if err := model.isStudentActive(); err != nil {
		return UpdateIEPResult{}, err
	}
	// TODO reimplement this
	// if model.isSame(command) {
	// 	return UpdateIEPResult{Skipped: true}, nil
	// }

	eventID := uuidv7.NewString()
	event := NewIEPUpdatedEvent(
		eventID,
		command,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateIEPResult{}, err
	}
	return UpdateIEPResult{EventID: eventID}, nil
}

type updateStudentIEPContext struct {
	iepExists       bool
	iepArchived     bool
	iepDeleted      bool
	iep             IEPState
	studentExists   bool
	studentArchived bool
	studentDeleted  bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadUpdateStudentIEPContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentIEPID,
	studentID string,
) (
	*updateStudentIEPContext,
	error,
) {
	query := streamQuery(studentIEPID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateStudentIEPContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateStudentIEPContext) isIEPActive() error {
	if !m.iepExists || m.iepDeleted {
		return eventstore.ErrIEPNotActive
	}
	return nil
}

func (m *updateStudentIEPContext) isStudentActive() error {
	if !m.studentExists || m.studentDeleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *updateStudentIEPContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case se.EventStudentCreated:
		m.studentExists = true
	case se.EventStudentArchived:
		m.studentArchived = true
	case se.EventStudentDeleted:
		m.studentDeleted = true
	case EventIEPAddedToStudent:
		m.iepExists = true
		event := IEPAddedToStudentEvent{}
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep update handle add unmarshal", "err", err)
			return
		}
		m.iep = event.IEPState
	case EventIEPUpdated:
		event := IEPUpdatedEvent{}
		if err := json.Unmarshal([]byte(rawData), &event); err != nil {
			slog.Error("iep update handle update unmarshal", "err", err)
			return
		}
		m.iep = event.IEPState
	case EventIEPArchived:
		m.iepArchived = true
	case EventIEPDeleted:
		m.iepDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
