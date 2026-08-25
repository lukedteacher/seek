package events

import (
	"seek/internal/eventstore"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func streamQuery(iepID, studentID string) eventstore.Query {
	iepEventTypes := []eventType{
		EventIEPAddedToStudent,
		EventIEPRemovedFromStudent,
		EventIEPUpdated,
		EventIEPArchived,
		EventIEPDeleted,
	}
	studentEventTypes := []eventType{
		se.EventStudentCreated,
		se.EventStudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(iepEventTypes)+len(studentEventTypes))
	for _, eventType := range iepEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
				{Key: FieldIEPScopeID, Value: iepID},
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

func studentStreamQuery(studentID string) eventstore.Query {
	iepEventTypes := []eventType{
		EventIEPAddedToStudent,
		EventIEPRemovedFromStudent,
		EventIEPUpdated,
		EventIEPArchived,
		EventIEPDeleted,
	}
	studentEventTypes := []eventType{
		se.EventStudentCreated,
		se.EventStudentArchived,
		se.EventStudentDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(iepEventTypes)+len(studentEventTypes))
	for _, eventType := range iepEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
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
