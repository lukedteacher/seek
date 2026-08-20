package events

import (
	"seek/internal/auth"
	"seek/internal/eventstore"
)

func extension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}

func registeredUserQuery(userRegisteredID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: auth.UserRegistered.String()},
				{Key: auth.UserRegisteredEventID, Value: userRegisteredID},
			}},
		},
	}
}

func profileUserEventQuery(eventType, userRegisteredID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: eventType},
				{Key: ProfileScopeUserRegisteredIDField, Value: userRegisteredID},
			}},
		},
	}
}

func combineQueries(queries ...eventstore.Query) eventstore.Query {
	combined := eventstore.Query{}
	for _, query := range queries {
		combined.Criteria = append(combined.Criteria, query.Criteria...)
	}
	return combined
}
