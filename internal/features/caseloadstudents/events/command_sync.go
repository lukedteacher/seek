package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
	pe "seek/internal/features/educators/events"
)

type SyncStudentsInEducatorCommand struct {
	EducatorID         string
	ProposedStudentIDs []string
	Metadata           CommandMetadata
}

type SyncStudentsInEducatorResult struct {
	Additions []AddStudentToCaseloadResult
	Removals  []RemoveStudentFromCaseloadResult
}

func SyncStudentsInEducatorCommandHandler(
	ctx context.Context,
	command SyncStudentsInEducatorCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*SyncStudentsInEducatorResult,
	error,
) {
	educator, err := loadSyncStudentsInCaseloadContext(ctx, retriever, command.EducatorID)
	if err != nil {
		return nil, fmt.Errorf("sync students in educator command handler: %w", err)
	}
	if err := educator.isEducatorActive(); err != nil {
		return nil, err
	}

	// build proposed map
	proposed := make(map[string]bool, len(command.ProposedStudentIDs))

	// check proposed against current and add students who are not present
	// also build the map for removals
	additions := []AddStudentToCaseloadResult{}
	for _, studentID := range command.ProposedStudentIDs {
		proposed[studentID] = true
		if _, ok := educator.students[studentID]; !ok {
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
				return &SyncStudentsInEducatorResult{}, nil
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveStudentFromCaseloadResult{}
	for studentID := range educator.students {
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
				return &SyncStudentsInEducatorResult{}, nil
			}
			removals = append(removals, *result)
		}
	}
	return &SyncStudentsInEducatorResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type educatorState struct {
	created  bool
	archived bool
	deleted  bool
}

type syncStudentsInEducatorContext struct {
	educator educatorState
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
		educator: educatorState{},
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
	case pe.EducatorCreated:
		m.educator.created = true
	case pe.EducatorArchived:
		m.educator.archived = true
	case pe.EducatorDeleted:
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
