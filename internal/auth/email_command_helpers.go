package auth

import "seek/internal/eventstore"

func userRegisteredQuery(userRegisteredID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: UserRegistered.String()},
				{Key: FieldUserRegisteredID, Value: userRegisteredID},
			}},
		},
	}
}

func userRegisteredByEmailQuery(emailHash string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: UserRegistered.String()},
				{Key: FieldUserRegisteredEmailHash, Value: emailHash},
			}},
		},
	}
}

func accountDeletionRequestedByUserQuery(userRegisteredID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: AccountDeletionRequested.String()},
				{Key: ScopeUserRegisteredEventIDField, Value: userRegisteredID},
			}},
		},
	}
}

func accountDeletedByRequestQuery(requestID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: AccountDeleted.String()},
				{Key: "scope.accountDeletionRequestedId", Value: requestID},
			}},
		},
	}
}

func accountDeletedByUserQuery(userRegisteredID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: AccountDeleted.String()},
				{Key: ScopeUserRegisteredEventIDField, Value: userRegisteredID},
			}},
		},
	}
}

func passwordResetRequestedQuery(requestID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: PasswordResetRequested.String()},
				{Key: PasswordResetRequestedEventID, Value: requestID},
			}},
		},
	}
}

func passwordResetEmailSentQuery(requestID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: PasswordResetEmailSent.String()},
				{Key: ScopePasswordResetRequestedEventIDField, Value: requestID},
			}},
		},
	}
}

func passwordResetCompletedQuery(requestID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: PasswordResetCompleted.String()},
				{Key: ScopePasswordResetRequestedEventIDField, Value: requestID},
			}},
		},
	}
}

func passwordChangedByUserQuery(userID string) eventstore.Query {
	return eventstore.Query{
		Criteria: []eventstore.Criterion{
			{Tags: []eventstore.Tag{
				{Key: eventTypeKey, Value: PasswordChanged.String()},
				{Key: ScopeUserRegisteredEventIDField, Value: userID},
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
