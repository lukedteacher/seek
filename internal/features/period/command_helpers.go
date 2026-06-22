package period

import (
	"errors"
	"strings"

	"seek/internal/eventstore"
)

type CommandMetadata = eventstore.CommandMetadata

func validateTitle(firstName string) (string, error) {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" || len(firstName) > 160 {
		return "", errors.New("student title must be between 1 and 160 characters")
	}
	return firstName, nil
}

func streamQuery(id string) eventstore.Query {
	criteria := make([]eventstore.Criterion, 0, 5)
	for _, eventType := range []string{PeriodCreated, PeriodRenamed, PeriodDeleted} {
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
