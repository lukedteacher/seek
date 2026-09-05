package events

import (
	"context"
	"encoding/json"
	"fmt"

	"seek/internal/eventstore"
)

type SyncStudentsInHomeroomCommand struct {
	HomeroomID         string
	ProposedStudentIDs []string
	Metadata           CommandMetadata
}

type SyncStudentsInHomeroomResult struct {
	Additions []AddStudentToHomeroomResult
	Removals  []RemoveStudentFromHomeroomResult
}

func SyncStudentsInHomeroomCommandHandler(
	ctx context.Context,
	cmd SyncStudentsInHomeroomCommand,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) (
	*SyncStudentsInHomeroomResult,
	error,
) {
	homeroom, err := loadSyncStudentsInHomeroomContext(ctx, saver, retriever, cmd.HomeroomID)
	if err != nil {
		return nil, fmt.Errorf("sync students in homeroom cmd handler: %w", err)
	}
	if err := homeroom.isHomeroomActive(); err != nil {
		return nil, err
	}

	// build proposed map
	proposed := make(map[string]bool, len(cmd.ProposedStudentIDs))
	// check proposed against current and add students who are not present
	// also build the map for removals
	additions := []AddStudentToHomeroomResult{}
	for _, studentID := range cmd.ProposedStudentIDs {
		if studentID == "" {
			continue
		}
		proposed[studentID] = true
		if _, ok := homeroom.students[studentID]; !ok {
			result, err := AddStudentToHomeroomCommandHandler(
				ctx,
				AddStudentToHomeroomCommand{
					HomeroomID: cmd.HomeroomID,
					StudentID:  studentID,
					Metadata:   cmd.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncStudentsInHomeroomResult{}, err
			}
			additions = append(additions, *result)
		}
	}

	// removals: current not in proposed
	removals := []RemoveStudentFromHomeroomResult{}
	for studentID := range homeroom.students {
		if !proposed[studentID] {
			result, err := RemoveStudentFromHomeroomCommandHandler(
				ctx,
				RemoveStudentFromHomeroomCommand{
					HomeroomID: cmd.HomeroomID,
					StudentID:  studentID,
					Metadata:   cmd.Metadata,
				},
				saver,
				retriever,
			)
			if err != nil {
				return &SyncStudentsInHomeroomResult{}, err
			}
			removals = append(removals, *result)
		}
	}
	return &SyncStudentsInHomeroomResult{
		Additions: additions,
		Removals:  removals,
	}, nil
}

type syncStudentsInHomeroomContext struct {
	homeroom homeroomState
	students map[string]struct{}
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func (m *syncStudentsInHomeroomContext) isHomeroomActive() error {
	if !m.homeroom.created || m.homeroom.archived || m.homeroom.deleted {
		return eventstore.ErrHomeroomNotActive
	}
	return nil
}

func loadSyncStudentsInHomeroomContext(
	ctx context.Context,
	_ eventstore.Saver,
	retriever eventstore.Retriever,
	homeroomID string,
) (
	*syncStudentsInHomeroomContext,
	error,
) {
	query := homeroomStudentSyncStreamQuery(homeroomID)
	events, err := retriever.GetEvents(
		ctx,
		eventstore.NoEventPosition,
		100,
		eventstore.Forward,
		query,
	)
	if err != nil {
		return nil, err
	}
	model := &syncStudentsInHomeroomContext{
		homeroom: homeroomState{},
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

func (m *syncStudentsInHomeroomContext) handle(resolved eventstore.ResolvedEvent) {
	data := resolved.Event.RawData
	switch resolved.Event.EventType {
	case EventHomeroomCreated:
		m.homeroom.created = true
	case EventHomeroomArchived:
		m.homeroom.archived = true
	case EventHomeroomDeleted:
		m.homeroom.deleted = true
	case EventStudentAddedToHomeroom:
		var event = &StudentAddedToHomeroomEvent{}
		_ = json.Unmarshal([]byte(data), event)
		m.students[event.StudentID] = struct{}{}
	case EventStudentRemovedFromHomeroom:
		var event = &StudentRemovedFromHomeroomEvent{}
		_ = json.Unmarshal([]byte(data), event)
		delete(m.students, event.StudentID)
	}
	if resolved.Position.After(m.position) {
		m.position = resolved.Position
	}
}
