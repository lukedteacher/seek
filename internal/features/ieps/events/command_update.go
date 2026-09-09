package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"seek/internal/eventstore"
	"seek/internal/features/ieps/models"
	studentEvents "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type UpdateIEPCommand struct {
	IEP      models.IEP
	Metadata CommandMetadata
}

type UpdateIEPResult struct {
	EventID string
	Skipped bool
}

func UpdateIEPCommandHandler(
	ctx context.Context,
	cmd UpdateIEPCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateIEPResult,
	error,
) {
	model, err := loadupdateIEPContext(
		ctx,
		retriever,
		cmd.IEP.ID,
		cmd.IEP.StudentID,
	)
	if err != nil {
		return UpdateIEPResult{}, err
	}
	if !model.isIEPActive() {
		return UpdateIEPResult{}, eventstore.ErrIEPNotActive
	}
	if !model.student.isActive() {
		return UpdateIEPResult{}, eventstore.ErrStudentNotActive
	}
	// TODO reimplement this
	// if model.isSame(cmd) {
	// 	return UpdateIEPResult{Skipped: true}, nil
	// }

	eventID := uuidv7.NewString()
	event := NewIEPUpdatedEvent(
		eventID,
		cmd,
		model.query,
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateIEPResult{}, err
	}
	return UpdateIEPResult{EventID: eventID}, nil
}

type updateIEPContext struct {
	iepExists   bool
	iepArchived bool
	iepDeleted  bool
	iep         models.IEP
	student     StudentState
	position    eventstore.Position
	events      []eventstore.ResolvedEvent
	query       eventstore.Query
}

func loadupdateIEPContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	iepID,
	studentID string,
) (
	*updateIEPContext,
	error,
) {
	query := StreamQuery(iepID, studentID)
	events, err := retriever.GetEvents(
		ctx,
		eventstore.NoEventPosition,
		100,
		eventstore.Forward,
		query,
	)
	if err != nil {
		return nil, err
	}

	model := &updateIEPContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateIEPContext) isIEPActive() bool {
	if m.iepExists || !m.iepArchived || !m.iepDeleted {
		return true
	}
	return false
}

func (m *updateIEPContext) handle(resolved eventstore.ResolvedEvent) {
	rawData := resolved.Event.RawData
	switch resolved.Event.EventType {
	case studentEvents.EventStudentCreated:
		m.student.created = true
	case studentEvents.EventStudentArchived:
		m.student.archived = true
	case studentEvents.EventStudentDeleted:
		m.student.deleted = true
	case EventIEPAddedToStudent:
		m.iepExists = true
		var flat IEPFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("iep update handle add unmarshal", "err", err)
			return
		}
		m.iep = NewModelFromFlat(flat)
	case EventIEPUpdated:
		var flat IEPFlat
		if err := json.Unmarshal([]byte(rawData), &flat); err != nil {
			slog.Error("iep update handle update unmarshal", "err", err)
			return
		}
		m.iep = NewModelFromFlat(flat)
	case EventIEPArchived:
		m.iepArchived = true
	case EventIEPDeleted:
		m.iepDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
