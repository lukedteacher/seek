package events

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreateIEPServiceCommand struct {
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

type CreateIEPServiceResult struct {
	EventID string
}

func CreateIEPServiceCommandHandler(
	ctx context.Context,
	command CreateIEPServiceCommand,
	saver eventstore.Saver,
) (
	CreateIEPServiceResult,
	error,
) {
	context, err := newCreateIEPServiceContext(command)
	if err != nil {
		return CreateIEPServiceResult{}, err
	}
	event := NewIEPServiceCreatedEvent(
		context.iepServiceID,
		context.studentID,
		context.serviceType,
		context.indirectMinutes,
		context.directMinutes,
		context.frequencyCount,
		context.frequencyType,
		context.location,
		context.startDate,
		context.endDate,
		context.provider,
		time.Now(),
		metadataWithQuery(command.Metadata, context.query),
	)
	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, eventstore.NoEventPosition, nil, context.query); err != nil {
		return CreateIEPServiceResult{}, err
	}
	return CreateIEPServiceResult{EventID: context.iepServiceID}, nil
}

type createIEPServiceContext struct {
	iepServiceID    string
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
	query           eventstore.Query
}

func newCreateIEPServiceContext(command CreateIEPServiceCommand) (*createIEPServiceContext, error) {
	iepServiceID := uuidv7.NewString()
	return &createIEPServiceContext{
		iepServiceID:    iepServiceID,
		studentID:       command.StudentID,
		serviceType:     command.ServiceType,
		indirectMinutes: command.IndirectMinutes,
		directMinutes:   command.DirectMinutes,
		frequencyCount:  command.FrequencyCount,
		frequencyType:   command.FrequencyType,
		location:        command.Location,
		startDate:       command.StartDate,
		endDate:         command.EndDate,
		provider:        command.Provider,
		query:           streamQuery(iepServiceID, command.StudentID),
	}, nil
}
