package schedule

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(id string) eventstore.Query {
	criteria := make([]eventstore.Criterion, 0, 2)
	for _, eventType := range []string{ScheduleCreated, ScheduleUpdated, ScheduleDeleted} {
		criteria = append(criteria, eventstore.Criterion{Tags: []eventstore.Tag{
			{Key: "eventType", Value: eventType},
			{Key: ScheduleScopeIDField, Value: id},
		}})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
