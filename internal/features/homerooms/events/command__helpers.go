package events

import (
	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func homeroomStreamQuery(homeroomID string) eventstore.Query {
	eventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func homeroomEducatorSyncStreamQuery(homeroomID string) eventstore.Query {
	eventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
		EventEducatorAddedToHomeroom,
		EventEducatorRemovedFromHomeroom,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func homeroomStudentSyncStreamQuery(homeroomID string) eventstore.Query {
	eventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
		EventStudentAddedToHomeroom,
		EventStudentRemovedFromHomeroom,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func homeroomEducatorStreamQuery(homeroomID, educatorID string) eventstore.Query {
	homeroomEventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(homeroomEventTypes))
	for _, eventType := range homeroomEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
			},
		})
	}
	homeroomEducatorEventTypes := []eventType{
		EventEducatorAddedToHomeroom,
		EventEducatorRemovedFromHomeroom,
	}
	for _, eventType := range homeroomEducatorEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
				{Key: FieldHomeroomEducatorScopeEducatorID, Value: educatorID},
			},
		})
	}

	return eventstore.Query{Criteria: criteria}
}

func homeroomStudentStreamQuery(homeroomID, studentID string) eventstore.Query {
	homeroomEventTypes := []eventType{
		EventHomeroomCreated,
		EventHomeroomUpdated,
		EventHomeroomArchived,
		EventHomeroomDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(homeroomEventTypes))
	for _, eventType := range homeroomEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
			},
		})
	}
	homeroomStudentEventTypes := []eventType{
		EventStudentAddedToHomeroom,
		EventStudentRemovedFromHomeroom,
	}
	for _, eventType := range homeroomStudentEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: FieldHomeroomScopeID, Value: homeroomID},
				{Key: FieldHomeroomStudentScopeStudentID, Value: studentID},
			},
		})
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}

func combineQueries(queries ...eventstore.Query) eventstore.Query {
	combined := eventstore.Query{}
	for _, query := range queries {
		combined.Criteria = append(combined.Criteria, query.Criteria...)
	}
	return combined
}
