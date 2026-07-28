package events

import (
	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(periodID, studentID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodCreated},
				{Key: pe.FieldPeriodCreatedEventID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodDeleted},
				{Key: pe.FieldPeriodDeletedEventID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.StudentCreated},
				{Key: se.StudentCreatedIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.StudentDeleted},
				{Key: se.StudentDeletedIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodStudentAdded},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.StudentScopeIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodStudentRemoved},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.StudentScopeIDField, Value: studentID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
