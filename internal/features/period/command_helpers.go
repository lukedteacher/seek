package period

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(periodID string) eventstore.Query {
	eventTypes := []string{
		PeriodCreated,
		PeriodUpdated,
		PeriodDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
				{Key: PeriodScopeIDField, Value: periodID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
