package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type UpdateIEPServiceCommand struct {
	IEPServiceID    string
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

type UpdateIEPServiceResult struct {
	EventID string
	Skipped bool
}

func UpdateIEPServiceCommandHandler(
	ctx context.Context,
	command UpdateIEPServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateIEPServiceResult,
	error,
) {
	model, err := loadUpdateIEPServiceContext(ctx, retriever, command.IEPServiceID, command.StudentID)
	if err != nil {
		return UpdateIEPServiceResult{}, err
	}
	if err := model.isServiceActive(); err != nil {
		return UpdateIEPServiceResult{}, err
	}
	if err := model.isStudentActive(); err != nil {
		return UpdateIEPServiceResult{}, err
	}
	if model.isSame(command) {
		return UpdateIEPServiceResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewIEPServiceUpdatedEvent(
		eventID,
		command.IEPServiceID,
		command.StudentID,
		command.ServiceName,
		command.ServiceType,
		command.IndirectMinutes,
		command.DirectMinutes,
		command.FrequencyCount,
		command.FrequencyType,
		command.LocationID,
		command.StartDate,
		command.EndDate,
		command.ProviderID,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateIEPServiceResult{}, err
	}
	return UpdateIEPServiceResult{EventID: eventID}, nil
}

type updateIEPServiceContext struct {
	serviceExists   bool
	serviceArchived bool
	serviceDeleted  bool
	studentExists   bool
	studentArchived bool
	studentDeleted  bool
	iepServiceID    string
	iepID           string
	studentID       string
	serviceName     string
	serviceType     string
	indirectMinutes int
	directMinutes   int
	frequencyCount  int
	frequencyType   string
	locationID      string
	startDate       string
	endDate         string
	providerID      string
	position        eventstore.Position
	events          []eventstore.ResolvedEvent
	query           eventstore.Query
}

func loadUpdateIEPServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	iepServiceID,
	studentID string,
) (
	*updateIEPServiceContext,
	error,
) {
	query := streamQuery(iepServiceID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateIEPServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateIEPServiceContext) isServiceActive() error {
	if !m.serviceExists || m.serviceDeleted {
		return eventstore.ErrServiceNotActive
	}
	return nil
}

func (m *updateIEPServiceContext) isStudentActive() error {
	if !m.studentExists || m.studentDeleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *updateIEPServiceContext) isSame(cmd UpdateIEPServiceCommand) bool {
	return m.studentID == cmd.StudentID &&
		m.serviceName == cmd.ServiceName &&
		m.serviceType == cmd.ServiceType &&
		m.indirectMinutes == cmd.IndirectMinutes &&
		m.directMinutes == cmd.DirectMinutes &&
		m.frequencyCount == cmd.FrequencyCount &&
		m.frequencyType == cmd.FrequencyType &&
		m.locationID == cmd.LocationID &&
		m.startDate == cmd.StartDate &&
		m.endDate == cmd.EndDate &&
		m.providerID == cmd.ProviderID
}

func (m *updateIEPServiceContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case se.EventStudentCreated:
		m.studentExists = true
	case se.EventStudentArchived:
		m.studentArchived = true
	case se.EventStudentDeleted:
		m.studentDeleted = true
	case EventServiceAddedToIEP:
		m.serviceExists = true
		m.serviceDeleted = false
		m.iepID, _ = data[FieldIEPServiceIEPID].(string)
		m.serviceName, _ = data[FieldIEPServiceServiceName].(string)
		m.serviceType, _ = data[FieldIEPServiceServiceType].(string)
		m.indirectMinutes = int(data[FieldIEPServiceIndirectMinutes].(float64))
		m.directMinutes = int(data[FieldIEPServiceDirectMinutes].(float64))
		m.frequencyCount = int(data[FieldIEPServiceFrequencyCount].(float64))
		m.frequencyType, _ = data[FieldIEPServiceFrequencyType].(string)
		m.locationID, _ = data[FieldIEPServiceLocationID].(string)
		m.startDate, _ = data[FieldIEPServiceStartDate].(string)
		m.endDate, _ = data[FieldIEPServiceEndDate].(string)
		m.providerID, _ = data[FieldIEPServiceProviderID].(string)
	case EventIEPServiceUpdated:
		m.iepID, _ = data[FieldIEPServiceIEPID].(string)
		m.serviceName, _ = data[FieldIEPServiceServiceName].(string)
		m.serviceType, _ = data[FieldIEPServiceServiceType].(string)
		m.indirectMinutes = int(data[FieldIEPServiceIndirectMinutes].(float64))
		m.directMinutes = int(data[FieldIEPServiceDirectMinutes].(float64))
		m.frequencyCount = int(data[FieldIEPServiceFrequencyCount].(float64))
		m.frequencyType, _ = data[FieldIEPServiceFrequencyType].(string)
		m.locationID, _ = data[FieldIEPServiceLocationID].(string)
		m.startDate, _ = data[FieldIEPServiceStartDate].(string)
		m.endDate, _ = data[FieldIEPServiceEndDate].(string)
		m.providerID, _ = data[FieldIEPServiceProviderID].(string)
	case EventIEPServiceDeleted:
		m.serviceDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
