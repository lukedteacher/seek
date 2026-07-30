package events

import (
	"context"
	"time"

	"seek/internal/eventstore"
	"seek/pkg/uuidv7"
)

type UpdateStudentCommand struct {
	StudentID   string
	MARSSID     string
	GivenName   string
	ChosenName  string
	FamilyName  string
	Grade       int
	Homeroom    string
	CaseManager string
	Metadata    CommandMetadata
}

type UpdateStudentResult struct {
	EventID string
	Skipped bool
}

func UpdateStudentCommandHandler(
	ctx context.Context,
	command UpdateStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	UpdateStudentResult,
	error,
) {
	model, err := loadUpdateStudentContext(ctx, retriever, command.StudentID)
	if err != nil {
		return UpdateStudentResult{}, err
	}
	if err := model.isActive(); err != nil {
		return UpdateStudentResult{}, err
	}
	if model.givenName == command.GivenName && model.chosenName == command.ChosenName && model.familyName == command.FamilyName && model.grade == command.Grade && model.homeroom == command.Homeroom && model.caseManager == command.CaseManager {
		return UpdateStudentResult{Skipped: true}, nil
	}
	eventID := uuidv7.NewString()
	event := NewStudentUpdatedEvent(
		eventID,
		command.StudentID,
		command.MARSSID,
		command.GivenName,
		command.ChosenName,
		command.FamilyName,
		command.Grade,
		command.Homeroom,
		command.CaseManager,
		time.Now(),
		metadataWithQuery(command.Metadata, model.query),
	)

	if _, err := saver.SaveEvents(ctx, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return UpdateStudentResult{}, err
	}
	return UpdateStudentResult{EventID: eventID}, nil
}

type updateStudentContext struct {
	created     bool
	archived    bool
	deleted     bool
	marssID     string
	givenName   string
	chosenName  string
	familyName  string
	grade       int
	homeroom    string
	caseManager string
	position    eventstore.Position
	events      []eventstore.ResolvedEvent
	query       eventstore.Query
}

func loadUpdateStudentContext(ctx context.Context, retriever eventstore.Retriever, studentID string) (*updateStudentContext, error) {
	query := StreamQuery(studentID)
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

func (m *updateStudentContext) isActive() error {
	if !m.created || m.archived || m.deleted {
		return eventstore.ErrNotFound
	}
	return nil
}

func (m *updateStudentContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.Data
	switch resolved.Event.EventType {
	case StudentCreated:
		m.created = true
		m.archived = false
		m.deleted = false
		m.givenName, _ = data[FieldStudentGivenName].(string)
		m.chosenName, _ = data[FieldStudentChosenName].(string)
		m.familyName, _ = data[FieldStudentFamilyName].(string)
		m.grade = int(data[FieldStudentGrade].(float64))
		m.homeroom, _ = data[FieldStudentHomeroom].(string)
		m.caseManager, _ = data[FieldStudentCaseManager].(string)
	case StudentUpdated:
		m.givenName, _ = data[FieldStudentGivenName].(string)
		m.chosenName, _ = data[FieldStudentChosenName].(string)
		m.familyName, _ = data[FieldStudentFamilyName].(string)
		m.grade = int(data[FieldStudentGrade].(float64))
		m.homeroom, _ = data[FieldStudentHomeroom].(string)
		m.caseManager, _ = data[FieldStudentCaseManager].(string)
	case StudentArchived:
		m.archived = true
	case StudentDeleted:
		m.deleted = true
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
