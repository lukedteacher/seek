package auth

import (
	"context"
	"strings"
	"time"

	"seek/internal/commandlimits"
	"seek/internal/eventstore"
	"seek/internal/features/users/models"
	"seek/pkg/uuidv7"
)

type RequestPasswordResetCommand struct {
	EmailAddress string
	Metadata     eventstore.CommandMetadata
}

type RequestPasswordResetResult struct {
	PasswordResetRequestedID string
	Token                    string
	ExpiresAt                time.Time
}

type PasswordResetUserReader interface {
	GetUserByEmailWithPassword(ctx context.Context, emailAddress string) (models.User, string, error)
}

func RequestPasswordResetCommandHandler(ctx context.Context, command RequestPasswordResetCommand, users PasswordResetUserReader, saver eventstore.Saver, retriever eventstore.Retriever, keys SubjectPiiKeyPort) (RequestPasswordResetResult, error) {
	if err := commandlimits.Assert(command); err != nil {
		return RequestPasswordResetResult{}, err
	}
	user, _, err := users.GetUserByEmailWithPassword(ctx, strings.ToLower(strings.TrimSpace(command.EmailAddress)))
	if err != nil {
		return RequestPasswordResetResult{}, nil
	}
	return requestPasswordReset(ctx, user, command.Metadata, saver, retriever, keys)
}

func requestPasswordReset(ctx context.Context, user models.User, metadata eventstore.CommandMetadata, saver eventstore.Saver, retriever eventstore.Retriever, keys SubjectPiiKeyPort) (RequestPasswordResetResult, error) {
	model, err := loadRequestPasswordResetContext(ctx, user, retriever)
	if err != nil {
		return RequestPasswordResetResult{}, err
	}

	token, err := randomToken(32)
	if err != nil {
		return RequestPasswordResetResult{}, err
	}
	requestID := uuidv7.NewString()
	expiresAt := time.Now().Add(30 * time.Minute)
	subjectKey, ok, err := keys.GetSubjectDataKey(ctx, user.UserRegisteredID)
	if err != nil {
		return RequestPasswordResetResult{}, err
	}
	if !ok {
		return RequestPasswordResetResult{}, eventstore.ErrSubjectKeyNotFound
	}
	event := NewPasswordResetRequestedEvent(requestID, user.Email, token, expiresAt, user.UserRegisteredID, subjectKey, nil)
	if _, err := eventstore.SaveCommandEvents(ctx, saver, metadata, []eventstore.DomainEvent{event}, model.position, model.events, model.query); err != nil {
		return RequestPasswordResetResult{}, err
	}
	return RequestPasswordResetResult{PasswordResetRequestedID: requestID, Token: token, ExpiresAt: expiresAt}, nil
}

type requestPasswordResetContext struct {
	position eventstore.Position
	events   []eventstore.ResolvedEvent
	query    eventstore.Query
}

func loadRequestPasswordResetContext(ctx context.Context, user models.User, retriever eventstore.Retriever) (*requestPasswordResetContext, error) {
	query := userRegisteredQuery(user.UserRegisteredID)
	latest, err := retriever.GetLatestByCriteria(ctx, query.Criteria)
	if err != nil {
		return nil, err
	}
	events := eventstore.EventsFromLatest(latest.Results)
	model := &requestPasswordResetContext{position: latest.ContextPosition, events: events, query: query}
	for _, event := range events {
		if event.Position.After(model.position) {
			model.position = event.Position
		}
	}
	return model, nil
}
