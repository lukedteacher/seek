package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/pkg/uuidv7"
)

type UpdatePeriodCommand struct {
	ID          string
	Title       string
	ServiceType sharedmodels.ServiceType
	StartTime   sharedmodels.TimeOnly
	Duration    int
	DaysBitmask sharedmodels.DaysBitmask
	Metadata    CommandMetadata
}

type UpdatePeriodResult struct {
	PeriodUpdatedID string
	Skipped         bool
}

func UpdatePeriodCommandHandler(
	ctx context.Context,
	command UpdatePeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdatePeriodResult,
	error,
) {
	model, err := loadUpdatePeriodContext(ctx, retriever, command.ID)
	if err != nil {
		return UpdatePeriodResult{}, err
	}
	if err := model.isActive(); err != nil {
		return UpdatePeriodResult{}, err
	}
	if model.isSame(command) {
		return UpdatePeriodResult{Skipped: true}, nil
	}
	eventID := uuidv7.NewString()
	event := NewPeriodUpdatedEvent(
		eventID,
		command.ID,
		command.Title,
		command.ServiceType,
		command.StartTime,
		command.Duration,
		command.DaysBitmask,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdatePeriodResult{}, err
	}
	return UpdatePeriodResult{PeriodUpdatedID: eventID}, nil
}

type updatePeriodContext struct {
	exists      bool
	archived    bool
	deleted     bool
	title       string
	serviceType sharedmodels.ServiceType
	startTime   sharedmodels.TimeOnly
	duration    int
	daysBitmask sharedmodels.DaysBitmask
	position    eventstore.Position
	events      []eventstore.ResolvedEvent
	query       eventstore.Query
}

func loadUpdatePeriodContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updatePeriodContext, error) {
	query := streamQuery(id)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updatePeriodContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updatePeriodContext) isActive() error {
	if !m.exists || m.archived || m.deleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func (m *updatePeriodContext) isSame(cmd UpdatePeriodCommand) bool {
	return m.title == cmd.Title &&
		m.serviceType == cmd.ServiceType &&
		m.startTime == cmd.StartTime &&
		m.duration == cmd.Duration &&
		m.daysBitmask == cmd.DaysBitmask
}

func (m *updatePeriodContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case EventPeriodCreated:
		m.exists = true
		m.title, _ = data[FieldPeriodTitle].(string)
		m.serviceType = sharedmodels.ServiceType(data[FieldPeriodServiceType].(string))
		m.startTime = parseDBTimeOnly(data[FieldPeriodStartTime].(string))
		m.duration = int(data[FieldPeriodDuration].(float64))
		m.daysBitmask = sharedmodels.DaysBitmask(int(data[FieldPeriodDaysBitmask].(float64)))
	case EventPeriodUpdated:
		m.title, _ = data[FieldPeriodTitle].(string)
		m.serviceType = sharedmodels.ServiceType(data[FieldPeriodServiceType].(string))
		m.startTime = parseDBTimeOnly(data[FieldPeriodStartTime].(string))
		m.duration = int(data[FieldPeriodDuration].(float64))
		m.daysBitmask = sharedmodels.DaysBitmask(int(data[FieldPeriodDaysBitmask].(float64)))
	case EventPeriodArchived:
		m.archived = true
	case EventPeriodDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
