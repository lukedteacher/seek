package events

import (
	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(serviceID, studentID string) eventstore.Query {
	studentEventTypes := []string{
		se.StudentCreated,
		se.StudentDeleted,
	}
	serviceEventTypes := []string{
		StudentServiceCreated,
		StudentServiceUpdated,
		StudentServiceDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(studentEventTypes) + len(serviceEventTypes))
	for _, eventType := range studentEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
				{Key: se.StudentScopeIDField, Value: studentID},
			},
		})
	}
	for _, eventType := range serviceEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
				{Key: se.StudentScopeIDField, Value: studentID},
				{Key: StudentServiceScopeIDField, Value: serviceID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
