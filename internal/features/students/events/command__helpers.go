package events

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func StreamQuery(studentID string) eventstore.Query {
	eventTypes := []string{
		StudentCreated,
		StudentUpdated,
		StudentArchived,
		StudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, 5)
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType},
				{Key: FieldStudentScopeID, Value: studentID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
