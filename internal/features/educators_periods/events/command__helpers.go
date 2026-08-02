package events

import (
	"seek/internal/eventstore"
	se "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(periodID, educatorID string) eventstore.Query {
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
				{Key: "eventType", Value: se.EducatorCreated},
				{Key: se.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.EducatorArchived},
				{Key: se.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.EducatorDeleted},
				{Key: se.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: EducatorRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: se.FieldEducatorScopeID, Value: educatorID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
