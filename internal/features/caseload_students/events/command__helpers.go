package events

import (
	"seek/internal/eventstore"
	ee "seek/internal/features/educators/events"
	se "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

var eventTypeKey = eventstore.EventTypeKey

func educatorStreamQuery(educatorID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorCreated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorUpdated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorArchived.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorDeleted.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func studentStreamQuery(studentID, educatorID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentCreated.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentArchived.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentDeleted.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorCreated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorUpdated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorArchived.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorDeleted.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func educatorStudentStreamQuery(educatorID, studentID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorCreated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorUpdated.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorArchived.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EventEducatorDeleted.String()},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentCreated.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentArchived.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.EventStudentDeleted.String()},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: ee.FieldScopeEducatorID, Value: educatorID},
				{Key: se.FieldScopeStudentID, Value: studentID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
