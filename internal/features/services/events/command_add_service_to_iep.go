package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	iepEvents "seek/internal/features/ieps/events"
	"seek/pkg/uuidv7"
)

type AddServiceToIEPCommand struct {
	IEPID           string
	StudentID       string
	ServiceName     string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	LocationID      string
	StartDate       string
	EndDate         string
	ProviderID      string
	Metadata        CommandMetadata
}

type AddServiceToIEPResult struct {
	EventID string
	Skipped bool
}

func AddServiceToIEPCommandHandler(
	ctx context.Context,
	cmd AddServiceToIEPCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*AddServiceToIEPResult,
	error,
) {
	model, err := loadAddServiceToIEPContext(
		ctx,
		retriever,
		cmd.IEPID,
		cmd.StudentID,
	)
	if err != nil {
		return nil, err
	}
	if err := model.isStudentActive(); err != nil {
		return nil, err
	}
	eventID := uuidv7.NewString()
	event := NewServiceAddedToStudentEvent(
		eventID,
		cmd,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(
		ctx,
		[]eventstore.DomainEvent{event},
		model.position,
		nil,
		model.query,
	); err != nil {
		return nil, err
	}
	return &AddServiceToIEPResult{EventID: eventID, Skipped: false}, nil
}

type addServiceToStudentContext struct {
	studentCreated  bool
	studentArchived bool
	studentDeleted  bool
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadAddServiceToIEPContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	iepID,
	studentID string,
) (
	*addServiceToStudentContext,
	error,
) {
	query := iepEvents.StreamQuery(iepID, studentID)
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
	model := &addServiceToStudentContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}
	// creates a model of the relevant context from past events
	for i := range events {
		model.handle(events[i])
	}
	return model, nil
}

func (m *addServiceToStudentContext) isStudentActive() error {
	if !m.studentCreated || m.studentArchived || m.studentDeleted {
		return eventstore.ErrPeriodNotFound
	}
	return nil
}

func (m *addServiceToStudentContext) handle(resolved eventstore.ResolvedEvent) {
	switch resolved.Event.EventType {
	case iepEvents.EventIEPAddedToStudent:
		m.studentCreated = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
