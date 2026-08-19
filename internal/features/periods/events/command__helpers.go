package events

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func streamQuery(periodID string) eventstore.Query {
	eventTypes := []string{
		PeriodCreated,
		PeriodUpdated,
		PeriodArchived,
		PeriodDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType},
				{Key: FieldPeriodScopeID, Value: periodID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
