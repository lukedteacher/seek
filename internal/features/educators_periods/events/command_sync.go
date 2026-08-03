package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
)

type SyncEducatorsInPeriodCommand struct {
	PeriodID            string
	ProposedEducatorIDs []string
	Metadata            CommandMetadata
}

type SyncEducatorsInPeriodResult struct {
	Additions []AddEducatorToPeriodResult
	Removals  []RemoveEducatorFromPeriodResult
}

func SyncEducatorsInPeriodCommandHandler(
	ctx context.Context,
	command SyncEducatorsInPeriodCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*SyncEducatorsInPeriodResult,
	error,
) {
	period, err := loadSyncEducatorsInPeriodContext(ctx, saver, retriever, command.PeriodID)
	if err != nil {
		return nil, fmt.Errorf("sync educators in period command handler: %w", err)
	}
	if err := period.isPeriodActive(); err != nil {
		return nil, err
	}

	// build proposed map
	proposed := make(map[string]bool, len(command.ProposedEducatorIDs))

	// check proposed against current and add educators who are not present
	// also build the map for removals
	additions := []AddEducatorToPeriodResult{}
	for _, educatorID := range command.ProposedEducatorIDs {
		proposed[educatorID] = true
		if _, ok := period.educators[educatorID]; !ok {
			result, err := AddEducatorToPeriodCommandHandler(
				ctx,
				AddEducatorToPeriodCommand{
					PeriodID:   command.PeriodID,
					EducatorID: educatorID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncEducatorsInPeriodResult{}, nil
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveEducatorFromPeriodResult{}
	for educatorID := range period.educators {
		if !proposed[educatorID] {
			result, err := RemoveEducatorFromPeriodCommandHandler(
				ctx,
				RemoveEducatorFromPeriodCommand{
					PeriodID:   command.PeriodID,
					EducatorID: educatorID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncEducatorsInPeriodResult{}, nil
			}
			removals = append(removals, *result)
		}
	}
	return &SyncEducatorsInPeriodResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type periodState struct {
	created  bool
	archived bool
	deleted  bool
}

type syncEducatorsInPeriodContext struct {
	period    periodState
	educators map[string]struct{}
	position  eventstore.Position
	events    []eventstore.ResolvedEvent
	query     eventstore.Query
}

func (m *syncEducatorsInPeriodContext) isPeriodActive() error {
	if !m.period.created || m.period.archived || m.period.deleted {
		return eventstore.ErrPeriodNotActive
	}
	return nil
}

func loadSyncEducatorsInPeriodContext(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	periodID string,
) (
	*syncEducatorsInPeriodContext,
	error,
) {
	query := periodStreamQuery(periodID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &syncEducatorsInPeriodContext{
		period:    periodState{},
		educators: make(map[string]struct{}),
		position:  eventstore.NoEventPosition,
		events:    events,
		query:     query,
	}

	for _, event := range events {
		model.handle(event)
	}
	return model, nil
}

func (m *syncEducatorsInPeriodContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
	case pe.PeriodCreated:
		m.period.created = true
	case pe.PeriodArchived:
		m.period.archived = true
	case pe.PeriodDeleted:
		m.period.deleted = true
	case EducatorAddedToPeriod:
		var event = &EducatorAddedToPeriodEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.educators[event.EducatorID] = struct{}{}
	case EducatorRemovedFromPeriod:
		var event = &EducatorRemovedFromPeriodEvent{}
		_ = json.Unmarshal([]byte(data), event)
		delete(m.educators, event.EducatorID)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
