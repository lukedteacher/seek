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
				{Key: eventTypeKey, Value: pe.EventPeriodCreated.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.EventPeriodUpdated.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.EventPeriodArchived.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.EventPeriodDeleted.String()},
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
				{Key: eventTypeKey, Value: pe.EventPeriodCreated.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.EventPeriodArchived.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: pe.EventPeriodDeleted.String()},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorCreated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorArchived.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorDeleted.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorAddedToPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: EducatorRemovedFromPeriod},
				{Key: pe.FieldPeriodScopeID, Value: periodID},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
