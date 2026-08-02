package auth

import (
	"context"
	"log/slog"
	"time"

	"seek/internal/eventstore"
	"seek/internal/protectedpii"
)

const AuthUserProjectionEventHandlerName = "auth_user_projection_event_handler"

type AuthUserProjectionWriter interface {
	CreateRegisteredUserAccount(ctx context.Context, registered RegisterUserResult) error
	MarkEmailVerified(ctx context.Context, userRegisteredID string) error
	UpdatePasswordByRegisteredID(ctx context.Context, userRegisteredID, passwordHash string) error
}

type AuthVerificationProjectionWriter interface {
	CreateEmailVerificationOTP(ctx context.Context, userRegisteredID, otpID, code string, expiresAt time.Time) error
	CreatePasswordReset(ctx context.Context, userRegisteredID, requestID, token string, expiresAt time.Time) error
}

type AuthUserProjectionEventHandler struct {
	global        *eventstore.GlobalEventHandler
	writer        AuthUserProjectionWriter
	verifications AuthVerificationProjectionWriter
	retriever     eventstore.Retriever
	keys          SubjectPiiKeyPort
}

func NewAuthUserProjectionEventHandler(
	subscriber eventstore.Subscriber,
	checkpointer eventstore.Checkpointer,
	retriever eventstore.Retriever,
	writer AuthUserProjectionWriter,
	verifications AuthVerificationProjectionWriter,
	keys SubjectPiiKeyPort,
	logger *slog.Logger) (*AuthUserProjectionEventHandler, error) {
	handler := &AuthUserProjectionEventHandler{retriever: retriever, writer: writer, verifications: verifications, keys: keys}
	global, err := eventstore.NewGlobalEventHandler(eventstore.GlobalEventHandlerConfig{
		Subscriber:      subscriber,
		Checkpointer:    checkpointer,
		Name:            AuthUserProjectionEventHandlerName,
		Query:           authUserProjectionEventHandlerQuery(),
		Logger:          logger,
		MaxEventRetries: -1,
		HandleEvent:     handler.handle,
	})
	if err != nil {
		return nil, err
	}
	handler.global = global
	return handler, nil
}

func (h *AuthUserProjectionEventHandler) StartSubscribing(ctx context.Context) error {
	return h.global.StartSubscribing(ctx)
}

func (h *AuthUserProjectionEventHandler) StopSubscribing() {
	h.global.StopSubscribing()
}

func (h *AuthUserProjectionEventHandler) handle(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	switch resolved.Event.EventType {
	case UserRegistered:
		return h.handleUserRegistered(ctx, resolved)
	case EmailVerificationOTPGenerated:
		return h.handleEmailVerificationOTPGenerated(ctx, resolved)
	case EmailVerificationOTPValidated:
		return h.handleEmailVerificationOTPValidated(ctx, resolved)
	case PasswordResetRequested:
		return h.handlePasswordResetRequested(ctx, resolved)
	case PasswordResetCompleted, PasswordChanged:
		return h.handlePasswordUpdated(ctx, resolved)
	default:
		return nil
	}
}

func (h *AuthUserProjectionEventHandler) handleUserRegistered(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	protector := protectedpii.FromEnv()
	userRegisteredID := stringValue(resolved.Event.Data[UserRegisteredIDField])
	subjectKey, ok, err := h.keys.GetSubjectDataKey(ctx, userRegisteredID)
	if err != nil {
		return err
	}
	if !ok {
		return eventstore.ErrNotFound
	}
	return h.writer.CreateRegisteredUserAccount(ctx, RegisterUserResult{
		EventID:      resolved.Event.EventID,
		Email:        protectedpii.MustDecryptEventStringWithDataKey(protector, subjectKey, resolved.Event.Data, UserRegisteredEmailField),
		PasswordHash: stringValue(resolved.Event.Data[UserRegisteredPasswordHashField]),
	})
}

func (h *AuthUserProjectionEventHandler) handleEmailVerificationOTPGenerated(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID := stringValue(eventstore.Scope(resolved.Event.Data)[UserRegisteredIDField])
	otpID := stringValue(resolved.Event.Data[EmailVerificationOTPGeneratedIDField])
	code := stringValue(resolved.Event.Data[EmailVerificationOTPCodeField])
	expiresAt, err := time.Parse(time.RFC3339, stringValue(resolved.Event.Data[EmailVerificationOTPExpiresAtField]))
	if userRegisteredID == "" || otpID == "" || code == "" || err != nil {
		return nil
	}
	return h.verifications.CreateEmailVerificationOTP(ctx, userRegisteredID, otpID, code, expiresAt)
}

func (h *AuthUserProjectionEventHandler) handleEmailVerificationOTPValidated(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID, _ := eventstore.Scope(resolved.Event.Data)[UserRegisteredIDField].(string)
	if userRegisteredID == "" {
		otpID, _ := eventstore.Scope(resolved.Event.Data)[EmailVerificationOTPGeneratedIDField].(string)
		generatedEvents, err := h.retriever.GetEvents(ctx, eventstore.NoEventPosition, 1, eventstore.Forward, emailVerificationOTPGeneratedQuery(otpID))
		if err != nil {
			return err
		}
		if len(generatedEvents) == 0 {
			return nil
		}
		userRegisteredID, _ = eventstore.Scope(generatedEvents[0].Event.Data)[UserRegisteredIDField].(string)
	}
	if userRegisteredID == "" {
		return nil
	}
	return h.writer.MarkEmailVerified(ctx, userRegisteredID)
}

func (h *AuthUserProjectionEventHandler) handlePasswordResetRequested(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID := stringValue(eventstore.Scope(resolved.Event.Data)[UserRegisteredIDField])
	requestID := stringValue(resolved.Event.Data[PasswordResetRequestedIDField])
	token := stringValue(resolved.Event.Data[PasswordResetRequestedTokenField])
	expiresAt, err := time.Parse(time.RFC3339, stringValue(resolved.Event.Data[PasswordResetRequestedExpiresAtField]))
	if userRegisteredID == "" || requestID == "" || token == "" || err != nil {
		return nil
	}
	return h.verifications.CreatePasswordReset(ctx, userRegisteredID, requestID, token, expiresAt)
}

func (h *AuthUserProjectionEventHandler) handlePasswordUpdated(ctx context.Context, resolved eventstore.ResolvedEvent) error {
	userRegisteredID := stringValue(eventstore.Scope(resolved.Event.Data)[UserRegisteredIDField])
	passwordHash := stringValue(resolved.Event.Data[PasswordChangedPasswordHashField])
	if resolved.Event.EventType == PasswordResetCompleted {
		passwordHash = stringValue(resolved.Event.Data[PasswordResetCompletedPasswordHashField])
	}
	if userRegisteredID == "" || passwordHash == "" {
		return nil
	}
	return h.writer.UpdatePasswordByRegisteredID(ctx, userRegisteredID, passwordHash)
}

func authUserProjectionEventHandlerQuery() eventstore.Query {
	eventTypes := []string{
		UserRegistered,
		EmailVerificationOTPGenerated,
		EmailVerificationOTPValidated,
		PasswordResetRequested,
		PasswordResetCompleted,
		PasswordChanged,
	}
	criteria := make([]eventstore.Criterion, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		criteria = append(criteria, eventstore.Criterion{
			Tags: []eventstore.Tag{
				{Key: "eventType", Value: eventType},
			},
		})
	}
	return eventstore.Query{Criteria: criteria}
}

func stringValue(value any) string {
	v, _ := value.(string)
	return v
}
