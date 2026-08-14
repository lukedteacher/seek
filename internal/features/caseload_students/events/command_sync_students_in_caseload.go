package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	ee "seek/internal/features/educators/events"
)

type SyncStudentsInCaseloadCommand struct {
	EducatorID         string
	ProposedStudentIDs []string
	Metadata           CommandMetadata
}

type SyncStudentsInCaseloadResult struct {
	Additions []AddStudentToCaseloadResult
	Removals  []RemoveStudentFromCaseloadResult
}

func SyncStudentsInCaseloadCommandHandler(
	ctx context.Context,
	command SyncStudentsInCaseloadCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	SyncStudentsInCaseloadResult,
	error,
) {
	model, err := loadSyncStudentsInCaseloadContext(ctx, retriever, command.EducatorID)
	if err != nil {
		return SyncStudentsInCaseloadResult{}, fmt.Errorf("sync students in educator command handler context: %w", err)
	}
	if err := model.isEducatorActive(); err != nil {
		return SyncStudentsInCaseloadResult{}, err
	}
	if !model.educator.isCaseManager {
		return SyncStudentsInCaseloadResult{}, fmt.Errorf("that educator is not a case manager")
	}

	// build proposed map
	proposed := make(map[string]bool, len(command.ProposedStudentIDs))

	// check proposed against current and add students who are not present
	// also build the map for removals
	additions := []AddStudentToCaseloadResult{}
	for _, studentID := range command.ProposedStudentIDs {
		proposed[studentID] = true
		if _, ok := model.students[studentID]; !ok {
			result, err := AddStudentToCaseloadCommandHandler(
				ctx,
				AddStudentToCaseloadCommand{
					EducatorID: command.EducatorID,
					StudentID:  studentID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return SyncStudentsInCaseloadResult{}, err
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveStudentFromCaseloadResult{}
	for studentID := range model.students {
		if !proposed[studentID] {
			result, err := RemoveStudentFromCaseloadCommandHandler(
				ctx,
				RemoveStudentFromCaseloadCommand{
					EducatorID: command.EducatorID,
					StudentID:  studentID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return SyncStudentsInCaseloadResult{}, err
			}
			removals = append(removals, *result)
		}
	}
	return SyncStudentsInCaseloadResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type syncStudentInCaseloadEducatorState struct {
	created       bool
	archived      bool
	deleted       bool
	isCaseManager bool
}

type syncStudentsInEducatorContext struct {
	educator syncStudentInCaseloadEducatorState
	students map[string]struct{}
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func (m *syncStudentsInEducatorContext) isEducatorActive() error {
	if !m.educator.created || m.educator.archived || m.educator.deleted {
		return eventstore.ErrEducatorNotActive
	}
	return nil
}

func loadSyncStudentsInCaseloadContext(
	ctx context.Context,
	retriever eventstore.Retriever,
	educatorID string,
) (
	*syncStudentsInEducatorContext,
	error,
) {
	query := educatorStreamQuery(educatorID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &syncStudentsInEducatorContext{
		educator: syncStudentInCaseloadEducatorState{},
		students: make(map[string]struct{}),
		position: eventstore.NoEventPosition,
		events:   events,
		query:    query,
	}

	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *syncStudentsInEducatorContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
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
		m.students[event.StudentID] = struct{}{}
	case StudentRemovedFromCaseload:
		var event = &StudentRemovedFromCaseloadEvent{}
		_ = json.Unmarshal([]byte(data), event)
		delete(m.students, event.StudentID)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
