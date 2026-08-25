package events

import (
	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func streamQuery(serviceID, studentID string) eventstore.Query {
	serviceEventTypes := []eventType{
		EventServiceAddedToIEP,
		EventServiceUpdated,
		EventServiceDeleted,
	}
	studentEventTypes := []eventType{
		se.EventStudentCreated,
		se.EventStudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(serviceEventTypes)+len(studentEventTypes))
	for _, eventType := range serviceEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
				{Key: FieldServiceScopeID, Value: serviceID},
			},
		})
	}
	for _, eventType := range studentEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
