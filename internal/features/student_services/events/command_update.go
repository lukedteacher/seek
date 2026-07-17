package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
	"seek/internal/uuidv7"
)

type UpdateStudentServiceCommand struct {
	ServiceID       string
	StudentID       string
	ServiceType     string
	IndirectMinutes int
	DirectMinutes   int
	FrequencyCount  int
	FrequencyType   string
	Location        string
	StartDate       string
	EndDate         string
	Provider        string
	Metadata        CommandMetadata
}

type UpdateStudentServiceResult struct {
	StudentServiceUpdatedID string
	Skipped                 bool
}

func UpdateStudentServiceCommandHandler(
	ctx context.Context,
	command UpdateStudentServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateStudentServiceResult,
	error,
) {
	model, err := loadUpdateStudentServiceContext(ctx, retriever, command.ServiceID, command.StudentID)
	if err != nil {
		return UpdateStudentServiceResult{}, err
	}
	if err := model.isServiceActive(); err != nil {
		return UpdateStudentServiceResult{}, err
	}
	if err := model.isStudentActive(); err != nil {
		return UpdateStudentServiceResult{}, err
	}
	if model.isSame(command) {
		return UpdateStudentServiceResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewStudentServiceUpdatedEvent(
		eventID,
		command.ServiceID,
		command.StudentID,
		command.ServiceType,
		command.IndirectMinutes,
		command.DirectMinutes,
		command.FrequencyCount,
		command.FrequencyType,
		command.Location,
		command.StartDate,
		command.EndDate,
		command.Provider,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateStudentServiceResult{}, err
	}
	return UpdateStudentServiceResult{StudentServiceUpdatedID: eventID}, nil
}

type updateStudentServiceContext struct {
	serviceExists   bool
	serviceDeleted  bool
	studentExists   bool
	studentDeleted  bool
	serviceID       string
	studentID       string
	serviceType     string
	indirectMinutes int
	directMinutes   int
	frequencyCount  int
	frequencyType   string
	location        string
	startDate       string
	endDate         string
	provider        string
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadUpdateStudentServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	serviceID,
	studentID string,
) (
	*updateStudentServiceContext,
	error,
) {
	query := streamQuery(serviceID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateStudentServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateStudentServiceContext) isServiceActive() error {
	if !m.serviceExists || m.serviceDeleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateStudentServiceContext) isStudentActive() error {
	if !m.studentExists || m.studentDeleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateStudentServiceContext) isSame(cmd UpdateStudentServiceCommand) bool {
	return m.studentID == cmd.StudentID &&
		m.serviceType == cmd.ServiceType &&
		m.indirectMinutes == cmd.IndirectMinutes &&
		m.directMinutes == cmd.DirectMinutes &&
		m.frequencyCount == cmd.FrequencyCount &&
		m.frequencyType == cmd.FrequencyType &&
		m.location == cmd.Location &&
		m.startDate == cmd.StartDate &&
		m.endDate == cmd.EndDate &&
		m.provider == cmd.Provider
}

func (m *updateStudentServiceContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case se.StudentCreated:
		m.studentExists = true
		m.studentDeleted = false
	case se.StudentDeleted:
		m.studentDeleted = true
	case StudentServiceCreated:
		m.serviceExists = true
		m.serviceDeleted = false
		m.studentID, _ = data[StudentServiceStudentIDField].(string)
		m.serviceType, _ = data[StudentServiceTypeField].(string)
		m.indirectMinutes = int(data[StudentServiceIndirectMinutesField].(float64))
		m.directMinutes = int(data[StudentServiceDirectMinutesField].(float64))
		m.frequencyCount = int(data[StudentServiceFrequencyCountField].(float64))
		m.frequencyType, _ = data[StudentServiceFrequencyTypeField].(string)
		m.location, _ = data[StudentServiceLocationField].(string)
		m.startDate, _ = data[StudentServiceStartDateField].(string)
		m.endDate, _ = data[StudentServiceEndDateField].(string)
		m.provider, _ = data[StudentServiceProviderField].(string)
	case StudentServiceUpdated:
		m.studentID, _ = data[StudentServiceStudentIDField].(string)
		m.serviceType, _ = data[StudentServiceTypeField].(string)
		m.indirectMinutes = int(data[StudentServiceIndirectMinutesField].(float64))
		m.directMinutes = int(data[StudentServiceDirectMinutesField].(float64))
		m.frequencyCount = int(data[StudentServiceFrequencyCountField].(float64))
		m.frequencyType, _ = data[StudentServiceFrequencyTypeField].(string)
		m.location, _ = data[StudentServiceLocationField].(string)
		m.startDate, _ = data[StudentServiceStartDateField].(string)
		m.endDate, _ = data[StudentServiceEndDateField].(string)
		m.provider, _ = data[StudentServiceProviderField].(string)
	case StudentServiceDeleted:
		m.serviceDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
