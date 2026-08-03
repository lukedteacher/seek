package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
)

type SyncStudentsInPeriodCommand struct {
	PeriodID           string
	ProposedStudentIDs []string
	Metadata           CommandMetadata
}

type SyncStudentsInPeriodResult struct {
	Additions []AddStudentToPeriodResult
	Removals  []RemoveStudentFromPeriodResult
}

func SyncStudentsInPeriodCommandHandler(
	ctx context.Context,
	command SyncStudentsInPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*SyncStudentsInPeriodResult,
	error,
) {
	period, err := loadSyncStudentsInPeriodContext(ctx, saver, retriever, command.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("sync students in period command handler: %w", err)
	}
	if err := period.isPeriodActive(); err != nil {
		return nil, err
	}

	// build proposed map
	proposed := make(map[string]bool, len(command.ProposedStudentIDs))

	// check proposed against current and add students who are not present
	// also build the map for removals
	additions := []AddStudentToPeriodResult{}
	for _, studentID := range command.ProposedStudentIDs {
		proposed[studentID] = true
		if _, ok := period.students[studentID]; !ok {
			result, err := AddStudentToPeriodCommandHandler(
				ctx,
				AddStudentToPeriodCommand{
					PeriodID:  command.PeriodID,
					StudentID: studentID,
					Metadata:  command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncStudentsInPeriodResult{}, nil
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveStudentFromPeriodResult{}
	for studentID := range period.students {
		if !proposed[studentID] {
			result, err := RemoveStudentFromPeriodCommandHandler(
				ctx,
				RemoveStudentFromPeriodCommand{
					PeriodID:  command.PeriodID,
					StudentID: studentID,
					Metadata:  command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncStudentsInPeriodResult{}, nil
			}
			removals = append(removals, *result)
		}
	}
	return &SyncStudentsInPeriodResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type periodState struct {
	created  bool
	archived bool
	deleted  bool
}

type syncStudentsInPeriodContext struct {
	period   periodState
	students map[string]struct{}
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func (m *syncStudentsInPeriodContext) isPeriodActive() error {
	if !m.period.created || m.period.archived || m.period.deleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func loadSyncStudentsInPeriodContext(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	periodID string,
) (
	*syncStudentsInPeriodContext,
	error,
) {
	query := periodStreamQuery(periodID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &syncStudentsInPeriodContext{
		period:   periodState{},
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

func (m *syncStudentsInPeriodContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		m.period.created = true
	case pe.PeriodArchived:
		m.period.archived = true
	case pe.PeriodDeleted:
		m.period.deleted = true
	case StudentAddedToPeriod:
		var event = &StudentAddedToPeriodEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.students[event.StudentID] = struct{}{}
	case StudentRemovedFromPeriod:
		var event = &StudentRemovedFromPeriodEvent{}
		_ = json.Unmarshal([]byte(data), event)
		delete(m.students, event.StudentID)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
