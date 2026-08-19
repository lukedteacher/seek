package events

import (
	"seek/internal/eventstore"
	ee "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func periodStreamQuery(periodID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodCreated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodUpdated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodArchived},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodDeleted},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorRemovedFromPeriod},
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
				{Key: eventTypeKey, Value: pe.PeriodCreated},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodArchived},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.PeriodDeleted},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorCreated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorArchived},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorDeleted},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorRemovedFromPeriod},
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
