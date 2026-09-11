package events

import (
	"seek/internal/auth"
	"seek/internal/eventstore"
	studentEvents "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func studentBookmarkStreamQuery(userID, studentID string) eventstore.Query {
	userEventTypes := []eventType{
		auth.UserRegistered,
		auth.AccountDeleted,
	}
	studentEventTypes := []eventType{
		studentEvents.EventStudentCreated,
		studentEvents.EventStudentArchived,
		studentEvents.EventStudentDeleted,
	}
	studentBookmarkEventTypes := []eventType{
		EventStudentBookmarkAdded,
		EventStudentBookmarkRemoved,
	}
	eventCount := len(userEventTypes) + len(studentEventTypes) + len(studentBookmarkEventTypes)
	criteria := make([]eventstore.Criterion, 0, eventCount)
	for _, eventType := range userEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: auth.UserRegisteredEventID, Value: userID},
			},
		})
	}
	for _, eventType := range studentEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: studentEvents.FieldScopeStudentID, Value: studentID},
			},
		})
	}
	for _, eventType := range studentBookmarkEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: auth.FieldUserRegisteredID, Value: userID},
				{Key: studentEvents.FieldStudentID, Value: studentID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}
func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
