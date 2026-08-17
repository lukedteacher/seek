package auth

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/protectedpii"
)

// user event types
const (
	UserRegistered           = "UserRegistered"
	PasswordResetRequested   = "PasswordResetRequested"
	PasswordResetEmailSent   = "PasswordResetEmailSent"
	PasswordResetCompleted   = "PasswordResetCompleted"
	PasswordChanged          = "PasswordChanged"
	AccountDeletionRequested = "AccountDeletionRequested"
	AccountDeleted           = "AccountDeleted"
	LoginAttemptRecorded     = "LoginAttemptRecorded"
)

// user event fields
const (
	UserRegisteredIDField                     = "user_registered_event_id"
	UserRegisteredEmailField                  = "email"
	UserRegisteredEmailHashField              = "email_hash"
	UserRegisteredPasswordHashField           = "password_hash"
	PasswordResetRequestedIDField             = "password_reset_requested_event_id"
	PasswordResetRequestedEmailField          = "email"
	PasswordResetRequestedEmailHashField      = "email_hash"
	PasswordResetRequestedTokenField          = "reset_token"
	PasswordResetRequestedTokenHashField      = "reset_token_hash"
	PasswordResetRequestedExpiresAtField      = "expires_at"
	PasswordResetEmailSentIDField             = "password_reset_email_sent_event_id"
	PasswordResetEmailSentAtField             = "sent_at"
	PasswordResetCompletedIDField             = "password_reset_completed_event_id"
	PasswordResetCompletedResetAtField        = "reset_at"
	PasswordResetCompletedPasswordHashField   = "password_hash"
	PasswordChangedIDField                    = "password_changed_id"
	PasswordChangedAtField                    = "changed_at"
	PasswordChangedPasswordHashField          = "password_hash"
	ScopeUserRegisteredIDField                = "scope.user_registered_event_id"
	ScopeEmailVerificationOTPGeneratedIDField = "scope.email_verification_otp_generated_event_id"
	ScopePasswordResetRequestedIDField        = "scope.password_reset_requested_event_id"
	AccountDeletionRequestedIDField           = "account_deletion_requested_event_id"
	AccountDeletionRequestedAtField           = "requested_at"
	AccountDeletedIDField                     = "account_deleted_event_id"
	AccountDeletedAtField                     = "deleted_at"
	LoginAttemptRecordedIDField               = "login_attempt_recorded_event_id"
	LoginAttemptIdentifierHashField           = "attempted_identifier_hash"
	LoginAttemptIPAddressHashField            = "ip_address_hash"
	LoginAttemptUserRegisteredIDField         = "user_registered_event_id"
)

type UserRegisteredEvent struct {
	EventID      string             `json:"user_registered_event_id"`
	Email        protectedpii.Value `json:"email"`
	EmailHash    string             `json:"email_hash"`
	PasswordHash string             `json:"password_hash"`
	Scope        map[string]any     `json:"scope"`
}

type PasswordResetRequestedEvent struct {
	PasswordResetRequestedID string             `json:"password_reset_requested_event_id"`
	Email                    protectedpii.Value `json:"email"`
	EmailHash                string             `json:"email_hash"`
	ResetToken               protectedpii.Value `json:"reset_token"`
	ResetTokenHash           string             `json:"reset_token_hash"`
	ExpiresAt                string             `json:"expires_at"`
	Scope                    UserScope          `json:"scope"`
}

type PasswordResetEmailSentEvent struct {
	PasswordResetEmailSentID string                      `json:"password_reset_email_sent_event_id"`
	SentAt                   string                      `json:"sent_at"`
	Scope                    PasswordResetRequestedScope `json:"scope"`
}

type PasswordResetCompletedEvent struct {
	PasswordResetCompletedID string                      `json:"password_reset_completed_event_id"`
	ResetAt                  string                      `json:"reset_at"`
	PasswordHash             string                      `json:"password_hash"`
	Scope                    PasswordResetCompletedScope `json:"scope"`
}

type PasswordChangedEvent struct {
	PasswordChangedID string    `json:"password_changed_event_id"`
	ChangedAt         string    `json:"changed_at"`
	PasswordHash      string    `json:"password_hash"`
	Scope             UserScope `json:"scope"`
}

type AccountDeletionRequestedEvent struct {
	AccountDeletionRequestedID string    `json:"account_deletion_requested_event_id"`
	RequestedAt                string    `json:"requested_at"`
	AuthUserID                 string    `json:"auth_user_id"`
	Scope                      UserScope `json:"scope"`
}

type AccountDeletedEvent struct {
	AccountDeletedID           string               `json:"account_deleted_event_id"`
	DeletedAt                  string               `json:"deleted_at"`
	AccountDeletionRequestedID string               `json:"account_deletion_requested_id"`
	AuthUserID                 string               `json:"auth_user_id"`
	Scope                      AccountDeletionScope `json:"scope"`
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
	UserRegisteredID string `json:"user_registered_event_id"`
}

type PasswordResetRequestedScope struct {
	PasswordResetRequestedID string `json:"password_reset_requested_event_id"`
}

type PasswordResetCompletedScope struct {
	PasswordResetRequestedID string `json:"password_reset_requested_event_id"`
	UserRegisteredID         string `json:"user_registered_id"`
}

type AccountDeletionScope struct {
	AccountDeletionRequestedID string `json:"account_deletion_requested_event_id"`
	UserRegisteredID           string `json:"user_registered_id"`
}

func NewUserRegisteredEvent(
	eventID,
	emailAddress,
	passwordHash string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: UserRegistered,
		Data: eventstore.MustData(UserRegisteredEvent{
			EventID:      eventID,
			Email:        protector.MustProtectWithDataKey(emailAddress, UserRegisteredEmailField, subjectKey),
			EmailHash:    protector.BlindIndex(UserRegisteredEmailField, emailAddress),
			PasswordHash: passwordHash,
			Scope:        map[string]any{},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetRequestedEvent(
	passwordResetRequestedID,
	emailAddress,
	resetToken string,
	expiresAt time.Time,
	userRegisteredID string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   passwordResetRequestedID,
		EventType: PasswordResetRequested,
		Data: eventstore.MustData(PasswordResetRequestedEvent{
			PasswordResetRequestedID: passwordResetRequestedID,
			Email:                    protector.MustProtectWithDataKey(emailAddress, PasswordResetRequestedEmailField, subjectKey),
			EmailHash:                protector.BlindIndex(PasswordResetRequestedEmailField, emailAddress),
			ResetToken:               protector.MustProtectWithDataKey(resetToken, PasswordResetRequestedTokenField, subjectKey),
			ResetTokenHash:           protector.SensitiveBlindIndex(PasswordResetRequestedTokenField, resetToken),
			ExpiresAt:                formatEventTime(expiresAt),
			Scope:                    UserScope{UserRegisteredID: userRegisteredID},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetEmailSentEvent(
	passwordResetEmailSentID string,
	sentAt time.Time,
	passwordResetRequestedID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordResetEmailSentID,
		EventType: PasswordResetEmailSent,
		Data: eventstore.MustData(PasswordResetEmailSentEvent{
			PasswordResetEmailSentID: passwordResetEmailSentID,
			SentAt:                   formatEventTime(sentAt),
			Scope:                    PasswordResetRequestedScope{PasswordResetRequestedID: passwordResetRequestedID},
		}),
		Metadata: metadata,
	}
}

func NewPasswordResetCompletedEvent(
	passwordResetCompletedID string,
	resetAt time.Time,
	passwordResetRequestedID,
	userRegisteredID,
	passwordHash string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordResetCompletedID,
		EventType: PasswordResetCompleted,
		Data: eventstore.MustData(PasswordResetCompletedEvent{
			PasswordResetCompletedID: passwordResetCompletedID,
			ResetAt:                  formatEventTime(resetAt),
			PasswordHash:             passwordHash,
			Scope: PasswordResetCompletedScope{
				PasswordResetRequestedID: passwordResetRequestedID,
				UserRegisteredID:         userRegisteredID,
			},
		}),
		Metadata: metadata,
	}
}

func NewPasswordChangedEvent(
	passwordChangedID string,
	changedAt time.Time,
	userRegisteredID,
	passwordHash string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   passwordChangedID,
		EventType: PasswordChanged,
		Data: eventstore.MustData(PasswordChangedEvent{
			PasswordChangedID: passwordChangedID,
			ChangedAt:         formatEventTime(changedAt),
			PasswordHash:      passwordHash,
			Scope:             UserScope{UserRegisteredID: userRegisteredID},
		}),
		Metadata: metadata,
	}
}

func NewAccountDeletionRequestedEvent(
	accountDeletionRequestedID string,
	requestedAt time.Time,
	authUserID,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   accountDeletionRequestedID,
		EventType: AccountDeletionRequested,
		Data: eventstore.MustData(AccountDeletionRequestedEvent{
			AccountDeletionRequestedID: accountDeletionRequestedID,
			RequestedAt:                formatEventTime(requestedAt),
			AuthUserID:                 authUserID,
			Scope:                      UserScope{UserRegisteredID: userRegisteredID},
		}),
		Metadata: metadata,
	}
}

func NewAccountDeletedEvent(
	accountDeletedID string,
	deletedAt time.Time,
	accountDeletionRequestedID,
	authUserID,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   accountDeletedID,
		EventType: AccountDeleted,
		Data: eventstore.MustData(AccountDeletedEvent{
			AccountDeletedID:           accountDeletedID,
			DeletedAt:                  formatEventTime(deletedAt),
			AccountDeletionRequestedID: accountDeletionRequestedID,
			AuthUserID:                 authUserID,
			Scope: AccountDeletionScope{
				AccountDeletionRequestedID: accountDeletionRequestedID,
				UserRegisteredID:           userRegisteredID,
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
