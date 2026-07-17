package events

import (
	"context"
	"time"

	"seek/internal/uuidv7"

	"seek/internal/eventstore"
)

type CreateStudentServiceCommand struct {
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

type CreateStudentServiceResult struct {
	ServiceID string
}

func CreateStudentServiceCommandHandler(
	ctx context.Context,
	command CreateStudentServiceCommand,
	saver eventstore.Saver,
) (
	CreateStudentServiceResult,
	error,
) {
	context, err := newCreateStudentServiceContext(command)
	if err != nil {
		return CreateStudentServiceResult{}, err
	}
	event := NewStudentServiceCreatedEvent(
		context.serviceID,
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
		return CreateStudentServiceResult{}, err
	}
	return CreateStudentServiceResult{ServiceID: context.serviceID}, nil
}

type createStudentServiceContext struct {
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
	query           eventstore.Query
}

func newCreateStudentServiceContext(command CreateStudentServiceCommand) (*createStudentServiceContext, error) {
	serviceID := uuidv7.NewString()
	return &createStudentServiceContext{
		serviceID:       serviceID,
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
		query:           streamQuery(command.StudentID, serviceID),
	}, nil
}
