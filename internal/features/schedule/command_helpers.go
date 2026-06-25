package schedule

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(id string) eventstore.Query {
	eventTypes := []string{
		ScheduleCreated,
		ScheduleUpdated,
		ScheduleDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
				{Key: ScheduleScopeIDField, Value: id},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
