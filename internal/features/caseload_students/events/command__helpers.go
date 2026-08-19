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
				{Key: eventTypeKey, Value: ee.EducatorCreated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorUpdated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorArchived},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorDeleted},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func studentStreamQuery(studentID, educatorID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentCreated},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentArchived},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentDeleted},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorCreated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorUpdated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorArchived},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorDeleted},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
	}
	return eventstore.Query{Criteria: criteria}
}

func educatorStudentStreamQuery(educatorID, studentID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorCreated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorUpdated},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorArchived},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: ee.EducatorDeleted},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentCreated},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentArchived},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: se.StudentDeleted},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentAddedToCaseload},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: StudentRemovedFromCaseload},
				{Key: ee.FieldEducatorScopeID, Value: educatorID},
				{Key: se.FieldStudentScopeID, Value: studentID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
