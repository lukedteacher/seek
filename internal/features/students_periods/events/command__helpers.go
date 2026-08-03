package events

import (
	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

func periodStreamQuery(periodID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodCreated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodUpdated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodArchived},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodDeleted},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: StudentAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: StudentRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func studentPeriodStreamQuery(periodID, studentID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodCreated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodArchived},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodDeleted},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.StudentCreated},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.StudentArchived},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.StudentDeleted},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: StudentAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: StudentRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func combineQueries(queries ...eventstore.Query) eventstore.Query {
	combined := eventstore.Query{}
	for _, query := range queries {
		combined.Criteria = append(combined.Criteria, query.Criteria...)
	}
	return combined
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
