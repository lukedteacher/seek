package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	ee "seek/internal/features/educators/events"
	se "seek/internal/features/students/events"
)

type SyncCaseManagerForStudentCommand struct {
	StudentID          string
	ProposedEducatorID string
	Metadata           CommandMetadata
}

type SyncCaseManagerForStudentResult struct {
	AddedTo     AddStudentToCaseloadResult
	RemovedFrom RemoveStudentFromCaseloadResult
	Skipped     bool
}

func SyncCaseManagerForStudentCommandHandler(
	ctx context.Context,
	command SyncCaseManagerForStudentCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	SyncCaseManagerForStudentResult,
	error,
) {
	model, err := loadSyncCaseManagerForStudentContext(ctx, retriever, command.StudentID, command.ProposedEducatorID)
	if err != nil {
		return SyncCaseManagerForStudentResult{}, fmt.Errorf("sync case manager for student command handler: %w", err)
	}
	if err := model.isStudentActive(); err != nil {
		return SyncCaseManagerForStudentResult{}, err
	}
	if err := model.isEducatorActive(); err != nil {
		return SyncCaseManagerForStudentResult{}, err
	}
	if !model.educator.isCaseManager {
		return SyncCaseManagerForStudentResult{}, fmt.Errorf("that educator is not a case manager")
	}

	// check proposed against current
	if model.student.caseManagerID == command.ProposedEducatorID {
		return SyncCaseManagerForStudentResult{Skipped: true}, nil
	}
	removedFromResult := &RemoveStudentFromCaseloadResult{}
	if model.student.caseManagerID != "" {
		removedFromResult, err = RemoveStudentFromCaseloadCommandHandler(
			ctx,
			RemoveStudentFromCaseloadCommand{
				EducatorID: model.student.caseManagerID,
				StudentID:  command.StudentID,
				Metadata:   command.Metadata,
			},
			saver,
			retriever,
		)
		if err != nil {
			return SyncCaseManagerForStudentResult{}, err
		}
	}
	addedToResult := &AddStudentToCaseloadResult{}
	if command.ProposedEducatorID != "" {
		addedToResult, err = AddStudentToCaseloadCommandHandler(
			ctx,
			AddStudentToCaseloadCommand{
				EducatorID: command.ProposedEducatorID,
				StudentID:  command.StudentID,
				Metadata:   command.Metadata,
			},
			saver,
			retriever,
		)
	}
	return SyncCaseManagerForStudentResult{
		AddedTo:     *addedToResult,
		RemovedFrom: *removedFromResult,
	}, nil
}

type syncCaseManagerForStudentStudentState struct {
	created       bool
	archived      bool
	deleted       bool
	caseManagerID string
}

type syncCaseManagerForStudentEducatorState struct {
	created       bool
	archived      bool
	deleted       bool
	isCaseManager bool
}

type syncCaseManagerForStudentContext struct {
	student  syncCaseManagerForStudentStudentState
	educator syncCaseManagerForStudentEducatorState
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func (m *syncCaseManagerForStudentContext) isStudentActive() error {
	if !m.student.created || m.student.archived || m.student.deleted {
		return eventstore.ErrStudentNotActive
	}
	return nil
}

func (m *syncCaseManagerForStudentContext) isEducatorActive() error {
	if !m.educator.created || m.educator.archived || m.educator.deleted {
		return eventstore.ErrEducatorNotActive
	}
	return nil
}

func loadSyncCaseManagerForStudentContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	studentID,
	educatorID string,
) (
	*syncCaseManagerForStudentContext,
	error,
) {
	query := studentStreamQuery(studentID, educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &syncCaseManagerForStudentContext{
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}

	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *syncCaseManagerForStudentContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
	case se.StudentCreated:
		m.student.created = true
	case se.StudentArchived:
		m.student.archived = true
	case se.StudentDeleted:
		m.student.deleted = true
	case ee.EducatorCreated:
		var event = &ee.EducatorCreatedEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.educator.created = true
		m.educator.isCaseManager = false
		for _, role := range event.Roles {
			if role == string(sharedmodels.EducatorRoleCaseManager) {
				m.educator.isCaseManager = true
			}
		}
	case ee.EducatorUpdated:
		var event = &ee.EducatorUpdatedEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.educator.isCaseManager = false
		for _, role := range event.Roles {
			if role == string(sharedmodels.EducatorRoleCaseManager) {
				m.educator.isCaseManager = true
			}
		}
	case ee.EducatorArchived:
		m.educator.archived = true
	case ee.EducatorDeleted:
		m.educator.deleted = true
	case StudentAddedToCaseload:
		var event = &StudentAddedToCaseloadEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.student.caseManagerID = event.EducatorID
	case StudentRemovedFromCaseload:
		var event = &StudentRemovedFromCaseloadEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.student.caseManagerID = ""
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
