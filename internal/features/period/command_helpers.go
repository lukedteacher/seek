package period

import (
	"errors"
	"strings"

	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > 160 {
		return "", errors.New("period title must be between 1 and 160 characters")
	}
	return title, nil
}

func validateStartTime(startTime string) (string, error) {
	startTime = strings.TrimSpace(startTime)
	if startTime == "" || len(startTime) > 160 {
		return "", errors.New("period start time must be between 1 and 160 characters")
	}
	return startTime, nil
}

func validateDuration(duration int64) (int64, error) {
	if duration < 1 || duration > 60 {
		return 0, errors.New("period duration must be between 1 and 60 minutes")
	}
	return duration, nil
}

func validateDays(days int64) (int64, error) {
	if days < 0 || days > 16 {
		return 0, errors.New("period days must be between 0 and 16")
	}
	return days, nil
}

func streamQuery(id string) eventstore.Query {
	criteria := make([]eventstore.Criterion, 0, 2)
	for _, eventType := range []string{PeriodCreated, PeriodUpdated, PeriodDeleted} {
		criteria = append(criteria, eventstore.Criterion{Tags: []eventstore.Tag{
			{Key: "eventType", Value: eventType},
			{Key: PeriodScopeIDField, Value: id},
		}})
	}
	return eventstore.Query{Criteria: criteria}
}

func metadataWithQuery(metadata CommandMetadata, query eventstore.Query) map[string]any {
	return eventstore.MergeMetadata(map[string]any{"query": eventstore.MustJSON(query)}, metadata)
}
