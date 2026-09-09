package events

import (
	"seek/internal/eventstore"
	iepEvents "seek/internal/features/ieps/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func serviceStreamQuery(serviceID, iepID string) eventstore.Query {
	serviceEventTypes := []eventType{
		EventServiceAddedToIEP,
		EventServiceUpdated,
		EventServiceDeleted,
	}
	iepEventTypes := []eventType{
		iepEvents.EventIEPAddedToStudent,
		iepEvents.EventIEPArchived,
		iepEvents.EventIEPDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(serviceEventTypes)+len(iepEventTypes))
	for _, eventType := range serviceEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: iepEvents.FieldIEPScopeID, Value: iepID},
				{Key: FieldServiceScopeID, Value: serviceID},
			},
		})
	}
	for _, eventType := range iepEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: iepEvents.FieldIEPScopeID, Value: iepID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func iepStreamQuery(iepID string) eventstore.Query {
	serviceEventTypes := []eventType{
		EventServiceAddedToIEP,
		EventServiceRemovedFromIEP,
		EventServiceUpdated,
		EventServiceArchived,
		EventServiceDeleted,
	}
	iepEventTypes := []eventType{
		iepEvents.EventIEPAddedToStudent,
		iepEvents.EventIEPArchived,
		iepEvents.EventIEPDeleted,
	}
	criteria := make([]eventstore.Criterion, 0, len(serviceEventTypes)+len(iepEventTypes))
	for _, eventType := range serviceEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: iepEvents.FieldIEPScopeID, Value: iepID},
			},
		})
	}
	for _, eventType := range iepEventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType.String()},
				{Key: iepEvents.FieldIEPScopeID, Value: iepID},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
