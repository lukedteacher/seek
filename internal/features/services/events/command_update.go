package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
	"seek/pkg/uuidv7"
)

type UpdateServiceCommand struct {
	ServiceID       string
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

type UpdateServiceResult struct {
	EventID string
	Skipped bool
}

func UpdateServiceCommandHandler(
	ctx context.Context,
	cmd UpdateServiceCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateServiceResult,
	error,
) {
	model, err := loadUpdateServiceContext(ctx, retriever, cmd.ServiceID, cmd.StudentID)
	if err != nil {
		return UpdateServiceResult{}, err
	}
	if err := model.isServiceActive(); err != nil {
		return UpdateServiceResult{}, err
	}
	if err := model.isStudentActive(); err != nil {
		return UpdateServiceResult{}, err
	}
	if model.isSame(cmd) {
		return UpdateServiceResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewServiceUpdatedEvent(
		eventID,
		cmd.ServiceID,
		cmd.IEPID,
		cmd.StudentID,
		cmd.ServiceName,
		cmd.ServiceType,
		cmd.IndirectMinutes,
		cmd.DirectMinutes,
		cmd.FrequencyCount,
		cmd.FrequencyType,
		cmd.LocationID,
		cmd.StartDate,
		cmd.EndDate,
		cmd.ProviderID,
		time.Now(),
		metadataWithQuery(cmd.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateServiceResult{}, err
	}
	return UpdateServiceResult{EventID: eventID}, nil
}

type updateServiceContext struct {
	serviceExists   bool
	serviceArchived bool
	serviceDeleted  bool
	studentExists   bool
	studentArchived bool
	studentDeleted  bool
	serviceID       string
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

func loadUpdateServiceContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	serviceID,
	studentID string,
) (
	*updateServiceContext,
	error,
) {
	query := streamQuery(serviceID, studentID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}

	model := &updateServiceContext{position: eventstore.NoEventPosition, events: events, query: query}
	for _, event := range events {
		model.handle(event)
	}

	return model, nil
}

func (m *updateServiceContext) isServiceActive() error {
	if !m.serviceExists || m.serviceDeleted {
		return eventstore.ErrServiceNotActive
	}
	return nil
}

func (m *updateServiceContext) isStudentActive() error {
	if !m.studentExists || m.studentDeleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *updateServiceContext) isSame(cmd UpdateServiceCommand) bool {
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

func (m *updateServiceContext) handle(resolved eventstore.ResolvedEvent) {
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
		m.iepID, _ = data[FieldServiceIEPID].(string)
		m.serviceName, _ = data[FieldServiceServiceName].(string)
		m.serviceType, _ = data[FieldServiceServiceType].(string)
		m.indirectMinutes = int(data[FieldServiceIndirectMinutes].(float64))
		m.directMinutes = int(data[FieldServiceDirectMinutes].(float64))
		m.frequencyCount = int(data[FieldServiceFrequencyCount].(float64))
		m.frequencyType, _ = data[FieldServiceFrequencyType].(string)
		m.locationID, _ = data[FieldServiceLocationID].(string)
		m.startDate, _ = data[FieldServiceStartDate].(string)
		m.endDate, _ = data[FieldServiceEndDate].(string)
		m.providerID, _ = data[FieldServiceProviderID].(string)
	case EventServiceUpdated:
		m.iepID, _ = data[FieldServiceIEPID].(string)
		m.serviceName, _ = data[FieldServiceServiceName].(string)
		m.serviceType, _ = data[FieldServiceServiceType].(string)
		m.indirectMinutes = int(data[FieldServiceIndirectMinutes].(float64))
		m.directMinutes = int(data[FieldServiceDirectMinutes].(float64))
		m.frequencyCount = int(data[FieldServiceFrequencyCount].(float64))
		m.frequencyType, _ = data[FieldServiceFrequencyType].(string)
		m.locationID, _ = data[FieldServiceLocationID].(string)
		m.startDate, _ = data[FieldServiceStartDate].(string)
		m.endDate, _ = data[FieldServiceEndDate].(string)
		m.providerID, _ = data[FieldServiceProviderID].(string)
	case EventServiceDeleted:
		m.serviceDeleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
