package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
)

type SyncEducatorsInHomeroomCommand struct {
	HomeroomID          string
	ProposedEducatorIDs []string
	Metadata            CommandMetadata
}

type SyncEducatorsInHomeroomResult struct {
	Additions []AddEducatorToHomeroomResult
	Removals  []RemoveEducatorFromHomeroomResult
}

func SyncEducatorsInHomeroomCommandHandler(
	ctx context.Context,
	command SyncEducatorsInHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*SyncEducatorsInHomeroomResult,
	error,
) {
	homeroom, err := loadSyncEducatorsInHomeroomContext(ctx, saver, retriever, command.HomeroomID)
	if err != nil {
		return nil, fmt.Errorf("sync educators in homeroom command handler: %w", err)
	}
	if err := homeroom.isHomeroomActive(); err != nil {
		return nil, err
	}

	// build proposed map
	proposed := make(map[string]bool, len(command.ProposedEducatorIDs))
	// check proposed against current and add educators who are not present
	// also build the map for removals
	additions := []AddEducatorToHomeroomResult{}
	for _, educatorID := range command.ProposedEducatorIDs {
		if educatorID == "" {
			continue
		}
		proposed[educatorID] = true
		if _, ok := homeroom.educators[educatorID]; !ok {
			result, err := AddEducatorToHomeroomCommandHandler(
				ctx,
				AddEducatorToHomeroomCommand{
					HomeroomID: command.HomeroomID,
					EducatorID: educatorID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncEducatorsInHomeroomResult{}, err
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveEducatorFromHomeroomResult{}
	for educatorID := range homeroom.educators {
		if !proposed[educatorID] {
			result, err := RemoveEducatorFromHomeroomCommandHandler(
				ctx,
				RemoveEducatorFromHomeroomCommand{
					HomeroomID: command.HomeroomID,
					EducatorID: educatorID,
					Metadata:   command.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncEducatorsInHomeroomResult{}, err
			}
			removals = append(removals, *result)
		}
	}
	return &SyncEducatorsInHomeroomResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type syncEducatorsInHomeroomContext struct {
	homeroom  homeroomState
	educators map[string]struct{}
	position  eventstore.Position
	events    []eventstore.ResolvedEvent
	query     eventstore.Query
}

func (m *syncEducatorsInHomeroomContext) isHomeroomActive() error {
	if !m.homeroom.created || m.homeroom.archived || m.homeroom.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func loadSyncEducatorsInHomeroomContext(
	ctx context.Context,
	_ eventstore.Saver,
	retriever eventstore.Retriever,
	homeroomID string,
) (
	*syncEducatorsInHomeroomContext,
	error,
) {
	query := homeroomEducatorSyncStreamQuery(homeroomID)
	events, err := retriever.GetEvents(ctx, eventstore.NoEventPosition, 100, eventstore.Forward, query)
	if err != nil {
		return nil, err
	}
	model := &syncEducatorsInHomeroomContext{
		homeroom:  homeroomState{},
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

func (m *syncEducatorsInHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.homeroom.created = true
	case EventHomeroomArchived:
		m.homeroom.archived = true
	case EventHomeroomDeleted:
		m.homeroom.deleted = true
	case EventEducatorAddedToHomeroom:
		var event = &EducatorAddedToHomeroomEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.educators[event.EducatorID] = struct{}{}
	case EventEducatorRemovedFromHomeroom:
		var event = &EducatorRemovedFromHomeroomEvent{}
		_ = json.Unmarshal([]byte(data), event)
		delete(m.educators, event.EducatorID)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
