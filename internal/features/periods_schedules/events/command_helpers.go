package events

import (
	"seek/internal/eventstore"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/schedules/events"
)

type CommandMetadata = eventstore.CommandMetadata

func streamQuery(periodID, scheduleID string) eventstore.Query {
	criteria := []eventstore.Criterion{
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodCreated},
				{Key: pe.PeriodCreatedIDField, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: pe.PeriodDeleted},
				{Key: pe.PeriodDeletedIDField, Value: periodID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.ScheduleCreated},
				{Key: se.ScheduleCreatedIDField, Value: scheduleID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: se.ScheduleDeleted},
				{Key: se.ScheduleDeletedIDField, Value: scheduleID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodScheduleAdded},
				{Key: pe.PeriodIDField, Value: periodID},
				{Key: se.ScheduleIDField, Value: scheduleID},
			},
		},
		{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: PeriodScheduleRemoved},
				{Key: pe.PeriodIDField, Value: periodID},
				{Key: se.ScheduleIDField, Value: scheduleID},
			},
		},
	}

	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
