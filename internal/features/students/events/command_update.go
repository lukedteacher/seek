package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/internal/uuidv7"
)

type UpdateStudentCommand struct {
	UserRegisteredID string
	Id               string
	FirstName        string
	ChosenName       string
	LastName         string
	Grade            int64
	Homeroom         string
	CaseManager      string
	Metadata         CommandMetadata
}

type UpdateStudentResult struct {
	StudentUpdatedID string
	Skipped          bool
}

func UpdateStudentCommandHandler(ctx context.Context, command UpdateStudentCommand, saver eventstore.Saver, retriever eventstore.Retriever) (UpdateStudentResult, error) {
	model, err := loadUpdateStudentContext(ctx, retriever, command.Id)
	if err != nil {
		return UpdateStudentResult{}, err
	}
	if err := model.requireActive(); err != nil {
		return UpdateStudentResult{}, err
	}
	if model.firstName == command.FirstName && model.chosenName == command.ChosenName && model.lastName == command.LastName && model.grade == command.Grade && model.homeroom == command.Homeroom && model.caseManager == command.CaseManager {
		return UpdateStudentResult{Skipped: true}, nil
	}

	eventID := uuidv7.NewString()
	event := NewStudentUpdatedEvent(eventID, command.Id, command.FirstName, command.ChosenName, command.LastName, command.Grade, command.Homeroom, command.CaseManager, time.Now(), metadataWithQuery(command.Metadata, model.query))

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateStudentResult{}, err
	}
	return UpdateStudentResult{StudentUpdatedID: eventID}, nil
}

type updateStudentContext struct {
	exists      bool
	deleted     bool
	firstName   string
	chosenName  string
	lastName    string
	grade       int64
	homeroom    string
	caseManager string
	position    eventstore.Position
	events      []eventstore.ResolvedEvent
	query       eventstore.Query
}

func loadUpdateStudentContext(ctx context.Context, retriever eventstore.Retriever, id string) (*updateStudentContext, error) {
	query := streamQuery(id)
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

func (m *updateStudentContext) requireActive() error {
	if !m.exists || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateStudentContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case StudentCreated:
		m.exists = true
		m.deleted = false
		m.firstName, _ = data[StudentFirstNameField].(string)
		m.chosenName, _ = data[StudentChosenNameField].(string)
		m.lastName, _ = data[StudentLastNameField].(string)
		m.grade = int64(data[StudentGradeField].(float64))
		m.homeroom, _ = data[StudentHomeroomField].(string)
		m.caseManager, _ = data[StudentCaseManagerField].(string)
	case StudentUpdated:
		m.firstName, _ = data[StudentFirstNameField].(string)
		m.chosenName, _ = data[StudentChosenNameField].(string)
		m.lastName, _ = data[StudentLastNameField].(string)
		m.grade = int64(data[StudentGradeField].(float64))
		m.homeroom, _ = data[StudentHomeroomField].(string)
		m.caseManager, _ = data[StudentCaseManagerField].(string)
	case StudentDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
