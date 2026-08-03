package events

import (
	"seek/internal/eventstore"
	ee "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
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
				{Key: "eventType", Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: EducatorRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func educatorPeriodStreamQuery(periodID, educatorID string) eventstore.Query {
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
				{Key: "eventType", Value: ee.EducatorCreated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: ee.EducatorArchived},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: ee.EducatorDeleted},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: EducatorRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
