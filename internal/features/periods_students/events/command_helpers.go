package events

import (
	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	student "seek/internal/features/students/events"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(periodID, studentID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: period.PeriodCreated},
				{Key: period.PeriodCreatedIDField, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: period.PeriodDeleted},
				{Key: period.PeriodDeletedIDField, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: student.StudentCreated},
				{Key: student.StudentCreatedIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: student.StudentDeleted},
				{Key: student.StudentDeletedIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodStudentAdded},
				{Key: period.PeriodIDField, Value: periodID},
				{Key: student.StudentIDField, Value: studentID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodStudentRemoved},
				{Key: period.PeriodIDField, Value: periodID},
				{Key: student.StudentIDField, Value: studentID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
