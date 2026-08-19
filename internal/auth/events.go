package auth

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/protectedpii"
)

type eventType = eventstore.EventType

var eventTypeKey = eventstore.EventTypeKey

// user event types
const (
	UserRegistered           eventType = "user_registered_event"
	PasswordResetRequested   eventType = "password_reset_requested_event"
	PasswordResetEmailSent   eventType = "password_reset_email_sent"
	PasswordResetCompleted   eventType = "password_reset_completed"
	PasswordChanged          eventType = "password_changed"
	AccountDeletionRequested eventType = "account_deletion_requested"
	AccountDeleted           eventType = "account_deleted"
	LoginAttemptRecorded     eventType = "login_attempt_recorded"
)

// user event ids
const (
	UserRegisteredEventID           = "user_registered_event_id"
	PasswordResetRequestedEventID   = "password_reset_requested_event_id"
	PasswordResetEmailSentEventID   = "password_reset_email_sent_event_id"
	PasswordResetCompletedEventID   = "password_reset_completed_event_id"
	PasswordChangedEventID          = "password_changed_event_id"
	AccountDeletionRequestedEventID = "account_deletion_requested_event_id"
	AccountDeletedEventID           = "account_deleted_event_id"
	LoginAttemptRecordedEventID     = "login_attempt_recorded_event_id"
)

// user event fields
const (
	FieldUserRegisteredID                   = "user_id"
	FieldUserRegisteredEmail                = "email"
	FieldUserRegisteredEmailHash            = "email_hash"
	FieldUserRegisteredUsername             = "username"
	FieldUserRegisteredPasswordHash         = "password_hash"
	FieldUserRegisteredRegisteredAt         = "registered_at"
	FieldPasswordResetRequestedEmail        = "email"
	FieldPasswordResetRequestedEmailHash    = "email_hash"
	FieldPasswordResetRequestedToken        = "reset_token"
	FieldPasswordResetRequestedTokenHash    = "reset_token_hash"
	FieldPasswordResetRequestedExpiresAt    = "expires_at"
	FieldPasswordResetEmailSentAt           = "sent_at"
	FieldPasswordResetCompletedResetAt      = "reset_at"
	FieldPasswordResetCompletedPasswordHash = "password_hash"
	FieldPasswordChangedAt                  = "changed_at"
	FieldPasswordChangedPasswordHash        = "password_hash"
	FieldAccountDeletionRequestedAt         = "requested_at"
	FieldAccountDeletedAt                   = "deleted_at"
	FieldLoginAttemptIdentifierHash         = "attempted_identifier_hash"
	FieldLoginAttemptIPAddressHash          = "ip_address_hash"
	ScopeUserRegisteredEventIDField         = "scope." + UserRegisteredEventID
	ScopePasswordResetRequestedEventIDField = "scope." + PasswordResetRequestedEventID
)

type UserRegisteredEvent struct {
	UserRegisteredEventID string             `json:"user_registered_event_id"`
	Email                 protectedpii.Value `json:"email"`
	EmailHash             string             `json:"email_hash"`
	Username              string             `json:"username"`
	PasswordHash          string             `json:"password_hash"`
	RegisteredAt          string             `json:"registered_at"`
	Scope                 map[string]any     `json:"scope"`
}

type PasswordResetRequestedEvent struct {
	PasswordResetRequestedEventID string                      `json:"password_reset_requested_event_id"`
	Email                         protectedpii.Value          `json:"email"`
	EmailHash                     string                      `json:"email_hash"`
	ResetToken                    protectedpii.Value          `json:"reset_token"`
	ResetTokenHash                string                      `json:"reset_token_hash"`
	ExpiresAt                     string                      `json:"expires_at"`
	Scope                         PasswordResetRequestedScope `json:"scope"`
}

type PasswordResetEmailSentEvent struct {
	PasswordResetEmailSentEventID string                      `json:"password_reset_email_sent_event_id"`
	SentAt                        string                      `json:"sent_at"`
	Scope                         PasswordResetRequestedScope `json:"scope"`
}

type PasswordResetCompletedEvent struct {
	PasswordResetCompletedEventID string                      `json:"password_reset_completed_event_id"`
	ResetAt                       string                      `json:"reset_at"`
	PasswordHash                  string                      `json:"password_hash"`
	Scope                         PasswordResetCompletedScope `json:"scope"`
}

type PasswordChangedEvent struct {
	PasswordChangedEventID string    `json:"password_changed_event_id"`
	ChangedAt              string    `json:"changed_at"`
	PasswordHash           string    `json:"password_hash"`
	Scope                  UserScope `json:"scope"`
}

type AccountDeletionRequestedEvent struct {
	AccountDeletionRequestedEventID string    `json:"account_deletion_requested_event_id"`
	RequestedAt                     string    `json:"requested_at"`
	AuthUserID                      string    `json:"auth_user_id"`
	Scope                           UserScope `json:"scope"`
}

type AccountDeletedEvent struct {
	AccountDeletedEventID      string                        `json:"account_deleted_event_id"`
	DeletedAt                  string                        `json:"deleted_at"`
	AccountDeletionRequestedID string                        `json:"account_deletion_requested_id"`
	AuthUserID                 string                        `json:"auth_user_id"`
	Scope                      AccountDeletionRequestedScope `json:"scope"`
}

type LoginAttemptRecordedEvent struct {
	LoginAttemptRecordedID  string             `json:"login_attempt_recorded_event_id"`
	AttemptedIdentifier     protectedpii.Value `json:"attempted_identifier"`
	AttemptedIdentifierHash string             `json:"attempted_identifier_hash"`
	IPAddress               protectedpii.Value `json:"ip_address"`
	IPAddressHash           string             `json:"ip_address_hash"`
	UserRegisteredID        string             `json:"user_registered_event_id,omitempty"`
	Succeeded               bool               `json:"succeeded"`
	RecordedAt              string             `json:"recorded_at"`
}

type UserScope struct {
	UserRegisteredEventID string `json:"user_registered_event_id"`
}

type PasswordResetRequestedScope struct {
	PasswordResetRequestedEventID string `json:"password_reset_requested_event_id"`
}

type PasswordResetCompletedScope struct {
	PasswordResetRequestedEventID string `json:"password_reset_requested_event_id"`
	UserRegisteredEventID         string `json:"user_registered_event_id"`
}

type AccountDeletionRequestedScope struct {
	AccountDeletionRequestedEventID string `json:"account_deletion_requested_event_id"`
	UserRegisteredEventID           string `json:"user_registered_event_id"`
}

func NewUserRegisteredEvent(
	userRegisteredEventID,
	email,
	passwordHash string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   userRegisteredEventID,
		EventType: UserRegistered,
		Data: eventstore.MustData(UserRegisteredEvent{
			UserRegisteredEventID: userRegisteredEventID,
			Email:                 protector.MustProtectWithDataKey(email, FieldUserRegisteredEmail, subjectKey),
			EmailHash:             protector.BlindIndex(FieldUserRegisteredEmail, email),
			Username:              deriveUsername(email),
			PasswordHash:          passwordHash,
			Scope:                 map[string]any{},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetRequestedEvent(
	passwordResetRequestedEventID,
	email,
	resetToken string,
	expiresAt time.Time,
	userRegisteredEventID string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   passwordResetRequestedEventID,
		EventType: PasswordResetRequested,
		Data: eventstore.MustData(PasswordResetRequestedEvent{
			PasswordResetRequestedEventID: passwordResetRequestedEventID,
			Email:                         protector.MustProtectWithDataKey(email, FieldPasswordResetRequestedEmail, subjectKey),
			EmailHash:                     protector.BlindIndex(FieldPasswordResetRequestedEmail, email),
			ResetToken:                    protector.MustProtectWithDataKey(resetToken, FieldPasswordResetRequestedToken, subjectKey),
			ResetTokenHash:                protector.SensitiveBlindIndex(FieldPasswordResetRequestedToken, resetToken),
			ExpiresAt:                     formatEventTime(expiresAt),
			Scope:                         PasswordResetRequestedScope{PasswordResetRequestedEventID: passwordResetRequestedEventID},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetEmailSentEvent(
	passwordResetEmailSentEventID string,
	sentAt time.Time,
	passwordResetRequestedEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordResetEmailSentEventID,
		EventType: PasswordResetEmailSent,
		Data: eventstore.MustData(PasswordResetEmailSentEvent{
			PasswordResetEmailSentEventID: passwordResetEmailSentEventID,
			SentAt:                        formatEventTime(sentAt),
			Scope:                         PasswordResetRequestedScope{PasswordResetRequestedEventID: passwordResetRequestedEventID},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetCompletedEvent(
	passwordResetCompletedID string,
	resetAt time.Time,
	passwordResetRequestedID,
	userRegisteredEventID,
	passwordHash string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordResetCompletedID,
		EventType: PasswordResetCompleted,
		Data: eventstore.MustData(PasswordResetCompletedEvent{
			PasswordResetCompletedEventID: passwordResetCompletedID,
			ResetAt:                       formatEventTime(resetAt),
			PasswordHash:                  passwordHash,
			Scope: PasswordResetCompletedScope{
				PasswordResetRequestedEventID: passwordResetRequestedID,
				UserRegisteredEventID:         userRegisteredEventID,
			},
		}),
		Metadata: metadata,
	}
}

func NewPasswordChangedEvent(
	passwordChangedEventID string,
	changedAt time.Time,
	userRegisteredEventID,
	passwordHash string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordChangedEventID,
		EventType: PasswordChanged,
		Data: eventstore.MustData(PasswordChangedEvent{
			PasswordChangedEventID: passwordChangedEventID,
			ChangedAt:              formatEventTime(changedAt),
			PasswordHash:           passwordHash,
			Scope:                  UserScope{UserRegisteredEventID: userRegisteredEventID},
		}),
		Metadata: metadata,
	}
}

func NewAccountDeletionRequestedEvent(
	accountDeletionRequestedEventID string,
	requestedAt time.Time,
	authUserID,
	userRegisteredEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   accountDeletionRequestedEventID,
		EventType: AccountDeletionRequested,
		Data: eventstore.MustData(AccountDeletionRequestedEvent{
			AccountDeletionRequestedEventID: accountDeletionRequestedEventID,
			RequestedAt:                     formatEventTime(requestedAt),
			AuthUserID:                      authUserID,
			Scope:                           UserScope{UserRegisteredEventID: userRegisteredEventID},
		}),
		Metadata: metadata,
	}
}

func NewAccountDeletedEvent(
	accountDeletedEventID string,
	deletedAt time.Time,
	accountDeletionRequestedEventID,
	authUserID,
	userRegisteredEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   accountDeletedEventID,
		EventType: AccountDeleted,
		Data: eventstore.MustData(AccountDeletedEvent{
			AccountDeletedEventID:      accountDeletedEventID,
			DeletedAt:                  formatEventTime(deletedAt),
			AccountDeletionRequestedID: accountDeletionRequestedEventID,
			AuthUserID:                 authUserID,
			Scope: AccountDeletionRequestedScope{
				AccountDeletionRequestedEventID: accountDeletionRequestedEventID,
				UserRegisteredEventID:           userRegisteredEventID,
			},
		}),
		Metadata: metadata,
	}
}

func NewLoginAttemptRecordedEvent(
	loginAttemptRecordedID string,
	recordedAt time.Time,
	attemptedIdentifier,
	ipAddress,
	userRegisteredID string,
	succeeded bool,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   loginAttemptRecordedID,
		EventType: LoginAttemptRecorded,
		Data: eventstore.MustData(LoginAttemptRecordedEvent{
			LoginAttemptRecordedID:  loginAttemptRecordedID,
			AttemptedIdentifier:     protector.MustProtect(attemptedIdentifier),
			AttemptedIdentifierHash: protector.BlindIndex("attemptedIdentifier", attemptedIdentifier),
			IPAddress:               protector.MustProtect(ipAddress),
			IPAddressHash:           protector.SensitiveBlindIndex("ipAddress", ipAddress),
			UserRegisteredID:        userRegisteredID,
			Succeeded:               succeeded,
			RecordedAt:              formatEventTime(recordedAt),
		}),
		Metadata: metadata,
	}
}

func formatEventTime(value time.Time) string {
	return value.Format(time.RFC3339)
}
