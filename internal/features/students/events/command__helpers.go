package events

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(id string) eventstore.Query {
	eventTypes := []string{
		StudentCreated,
		StudentUpdated,
		StudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, 5)
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
				{Key: StudentScopeIDField, Value: id},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
