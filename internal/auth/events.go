package auth

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/protectedpii"
)

// EVENT NAMES
const (
	UserRegistered                = "UserRegistered"
	UserNameChanged               = "UserNameChanged"
	EmailVerificationOTPGenerated = "EmailVerificationOTPGenerated"
	EmailVerificationOTPValidated = "EmailVerificationOTPValidated"
	EmailVerificationOTPSent      = "EmailVerificationOTPSent"
	PasswordResetRequested        = "PasswordResetRequested"
	PasswordResetEmailSent        = "PasswordResetEmailSent"
	PasswordResetCompleted        = "PasswordResetCompleted"
	PasswordChanged               = "PasswordChanged"
	AccountDeletionRequested      = "AccountDeletionRequested"
	AccountDeleted                = "AccountDeleted"
	LoginAttemptRecorded          = "LoginAttemptRecorded"
)

// EVENT FIELDS
const (
	UserRegisteredIDField                     = "user_registered_id"
	UserRegisteredUsernameField               = "username"
	UserRegisteredUsernameHashField           = "username_hash"
	UserRegisteredEmailField                  = "email"
	UserRegisteredEmailHashField              = "email_hash"
	UserRegisteredFirstNameField              = "first_name"
	UserRegisteredLastNameField               = "last_name"
	UserRegisteredYearOfBirthField            = "year_of_birth"
	UserRegisteredPasswordHashField           = "password_hash"
	UserNameChangedIDField                    = "user_name_changed_event_id"
	UserNameChangedNameField                  = "name"
	UserNameChangedNameHashField              = "name_hash"
	UserNameChangedChangedAtField             = "changed_at"
	EmailVerificationOTPGeneratedIDField      = "email_verification_otp_generated_id"
	EmailVerificationOTPCodeField             = "otp_code"
	EmailVerificationOTPExpiresAtField        = "expires_at"
	EmailVerificationOTPValidatedIDField      = "email_verification_otp_validated_id"
	EmailVerificationOTPValidatedAtField      = "validated_at"
	EmailVerificationOTPSentIDField           = "email_verification_otp_sent_id"
	EmailVerificationOTPSentAtField           = "sent_at"
	PasswordResetRequestedIDField             = "password_reset_requested_id"
	PasswordResetRequestedEmailField          = "email"
	PasswordResetRequestedEmailHashField      = "email_hash"
	PasswordResetRequestedTokenField          = "reset_token"
	PasswordResetRequestedTokenHashField      = "reset_token_hash"
	PasswordResetRequestedExpiresAtField      = "expires_at"
	PasswordResetEmailSentIDField             = "password_reset_email_sent_id"
	PasswordResetEmailSentAtField             = "sent_at"
	PasswordResetCompletedIDField             = "password_reset_completed_id"
	PasswordResetCompletedResetAtField        = "reset_at"
	PasswordResetCompletedPasswordHashField   = "password_hash"
	PasswordChangedIDField                    = "password_changed_id"
	PasswordChangedAtField                    = "changed_at"
	PasswordChangedPasswordHashField          = "password_hash"
	ScopeUserRegisteredIDField                = "scope.user_registered_id"
	ScopeEmailVerificationOTPGeneratedIDField = "scope.email_verification_otp_generated_id"
	ScopePasswordResetRequestedIDField        = "scope.password_reset_requested_id"
	AccountDeletionRequestedIDField           = "account_deletion_requested_id"
	AccountDeletionRequestedAtField           = "requested_at"
	AccountDeletedIDField                     = "accountDeleted_id"
	AccountDeletedAtField                     = "deleted_at"
	LoginAttemptRecordedIDField               = "login_attempt_recorded_id"
	LoginAttemptIdentifierHashField           = "attempted_identifier_hash"
	LoginAttemptIPAddressHashField            = "ip_address_hash"
	LoginAttemptUserRegisteredIDField         = "user_registered_id"
)

type UserRegisteredEvent struct {
	UserRegisteredID string             `json:"user_registered_id"`
	Username         protectedpii.Value `json:"username"`
	UsernameHash     string             `json:"username_hash"`
	Email            protectedpii.Value `json:"email"`
	EmailHash        string             `json:"email_hash"`
	FirstName        protectedpii.Value `json:"first_ame"`
	LastName         protectedpii.Value `json:"last_name"`
	YearOfBirth      int                `json:"year_of_birth"`
	PasswordHash     string             `json:"password_hash"`
	Scope            map[string]any     `json:"scope"`
}

type UserNameChangedEvent struct {
	UserNameChangedID string              `json:"user_name_changed_id"`
	Name              protectedpii.Value  `json:"name"`
	NameHash          string              `json:"name_hash"`
	ChangedAt         string              `json:"changed_at"`
	Scope             UserRegisteredScope `json:"scope"`
}

type EmailVerificationOTPGeneratedEvent struct {
	EmailVerificationOTPGeneratedID string              `json:"email_verification_otp_generated_id"`
	OTPCode                         string              `json:"otp_code"`
	ExpiresAt                       string              `json:"expires_at"`
	Scope                           UserRegisteredScope `json:"scope"`
}

type EmailVerificationOTPValidatedEvent struct {
	EmailVerificationOTPValidatedID string                             `json:"email_verification_otp_validated_id"`
	ValidatedAt                     string                             `json:"validated_at"`
	Scope                           EmailVerificationOTPValidatedScope `json:"scope"`
}

type EmailVerificationOTPSentEvent struct {
	EmailVerificationOTPSentID string                             `json:"email_verification_otp_sent_id"`
	SentAt                     string                             `json:"sent_at"`
	Scope                      EmailVerificationOTPGeneratedScope `json:"scope"`
}

type PasswordResetRequestedEvent struct {
	PasswordResetRequestedID string              `json:"password_reset_requested_id"`
	Email                    protectedpii.Value  `json:"email"`
	EmailHash                string              `json:"email_hash"`
	ResetToken               protectedpii.Value  `json:"reset_token"`
	ResetTokenHash           string              `json:"reset_token_hash"`
	ExpiresAt                string              `json:"expires_at"`
	Scope                    UserRegisteredScope `json:"scope"`
}

type PasswordResetEmailSentEvent struct {
	PasswordResetEmailSentID string                      `json:"password_reset_email_sent_id"`
	SentAt                   string                      `json:"sent_at"`
	Scope                    PasswordResetRequestedScope `json:"scope"`
}

type PasswordResetCompletedEvent struct {
	PasswordResetCompletedID string                      `json:"password_reset_completed_id"`
	ResetAt                  string                      `json:"reset_at"`
	PasswordHash             string                      `json:"password_hash"`
	Scope                    PasswordResetCompletedScope `json:"scope"`
}

type PasswordChangedEvent struct {
	PasswordChangedID string              `json:"password_changed_id"`
	ChangedAt         string              `json:"changed_at"`
	PasswordHash      string              `json:"password_hash"`
	Scope             UserRegisteredScope `json:"scope"`
}

type AccountDeletionRequestedEvent struct {
	AccountDeletionRequestedID string              `json:"account_deletion_requested_id"`
	RequestedAt                string              `json:"requested_at"`
	AuthUserID                 string              `json:"auth_user_id"`
	Scope                      UserRegisteredScope `json:"scope"`
}

type AccountDeletedEvent struct {
	AccountDeletedID           string               `json:"account_deleted_id"`
	DeletedAt                  string               `json:"deleted_at"`
	AccountDeletionRequestedID string               `json:"account_deletion_requested_id"`
	AuthUserID                 string               `json:"auth_user_id"`
	Scope                      AccountDeletionScope `json:"scope"`
}

type LoginAttemptRecordedEvent struct {
	LoginAttemptRecordedID  string             `json:"login_attempt_recorded_id"`
	AttemptedIdentifier     protectedpii.Value `json:"attempted_identifier"`
	AttemptedIdentifierHash string             `json:"attempted_identifier_hash"`
	IPAddress               protectedpii.Value `json:"ip_address"`
	IPAddressHash           string             `json:"ip_address_hash"`
	UserRegisteredID        string             `json:"user_registered_id,omitempty"`
	Succeeded               bool               `json:"succeeded"`
	RecordedAt              string             `json:"recorded_at"`
}

type UserRegisteredScope struct {
	UserRegisteredID string `json:"user_registered_id"`
}

type EmailVerificationOTPGeneratedScope struct {
	EmailVerificationOTPGeneratedID string `json:"email_verification_otp_generated_id"`
}

type EmailVerificationOTPValidatedScope struct {
	EmailVerificationOTPGeneratedID string `json:"email_verification_otp_validated_id"`
	UserRegisteredID                string `json:"user_registered_id"`
}

type PasswordResetRequestedScope struct {
	PasswordResetRequestedID string `json:"password_reset_requested_id"`
}

type PasswordResetCompletedScope struct {
	PasswordResetRequestedID string `json:"password_reset_requested_id"`
	UserRegisteredID         string `json:"user_registered_id"`
}

type AccountDeletionScope struct {
	AccountDeletionRequestedID string `json:"account_deletion_requested_id"`
	UserRegisteredID           string `json:"user_registered_id"`
}

func NewUserRegisteredEvent(
	userRegisteredID,
	username,
	emailAddress,
	firstName,
	lastName string,
	yearOfBirth int,
	passwordHash string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   userRegisteredID,
		EventType: UserRegistered,
		Data: eventstore.MustData(UserRegisteredEvent{
			UserRegisteredID: userRegisteredID,
			Username:         protector.MustProtectWithDataKey(username, UserRegisteredUsernameField, subjectKey),
			UsernameHash:     protector.BlindIndex(UserRegisteredUsernameField, username),
			Email:            protector.MustProtectWithDataKey(emailAddress, UserRegisteredEmailField, subjectKey),
			EmailHash:        protector.BlindIndex(UserRegisteredEmailField, emailAddress),
			FirstName:        protector.MustProtectWithDataKey(firstName, UserRegisteredFirstNameField, subjectKey),
			LastName:         protector.MustProtectWithDataKey(lastName, UserRegisteredLastNameField, subjectKey),
			YearOfBirth:      yearOfBirth,
			PasswordHash:     passwordHash,
			Scope:            map[string]any{},
		}),
		Metadata: metadata,
	}
}

func NewUserNameChangedEvent(
	userNameChangedID,
	name string,
	changedAt time.Time,
	userRegisteredID string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	return eventstore.DomainEvent{
		EventID:   userNameChangedID,
		EventType: UserNameChanged,
		Data: eventstore.MustData(UserNameChangedEvent{
			UserNameChangedID: userNameChangedID,
			Name:              protector.MustProtectWithDataKey(name, UserNameChangedNameField, subjectKey),
			NameHash:          protector.BlindIndex(UserNameChangedNameField, name),
			ChangedAt:         formatEventTime(changedAt),
			Scope:             UserRegisteredScope{UserRegisteredID: userRegisteredID},
		}),
		Metadata: metadata,
	}
}

func NewEmailVerificationOTPGeneratedEvent(
	emailVerificationOTPGeneratedID,
	otpCode string,
	expiresAt time.Time,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   emailVerificationOTPGeneratedID,
		EventType: EmailVerificationOTPGenerated,
		Data: eventstore.MustData(EmailVerificationOTPGeneratedEvent{
			EmailVerificationOTPGeneratedID: emailVerificationOTPGeneratedID,
			OTPCode:                         otpCode,
			ExpiresAt:                       formatEventTime(expiresAt),
			Scope:                           UserRegisteredScope{UserRegisteredID: userRegisteredID},
		}),
		Metadata: metadata,
	}
}

func NewEmailVerificationOTPValidatedEvent(
	emailVerificationOTPValidatedID string,
	validatedAt time.Time,
	emailVerificationOTPGeneratedID,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   emailVerificationOTPValidatedID,
		EventType: EmailVerificationOTPValidated,
		Data: eventstore.MustData(EmailVerificationOTPValidatedEvent{
			EmailVerificationOTPValidatedID: emailVerificationOTPValidatedID,
			ValidatedAt:                     formatEventTime(validatedAt),
			Scope: EmailVerificationOTPValidatedScope{
				EmailVerificationOTPGeneratedID: emailVerificationOTPGeneratedID,
				UserRegisteredID:                userRegisteredID,
			},
		}),
		Metadata: metadata,
	}
}

func NewEmailVerificationOTPSentEvent(
	emailVerificationOTPSentID string,
	sentAt time.Time,
	emailVerificationOTPGeneratedID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	return eventstore.DomainEvent{
		EventID:   emailVerificationOTPSentID,
		EventType: EmailVerificationOTPSent,
		Data: eventstore.MustData(EmailVerificationOTPSentEvent{
			EmailVerificationOTPSentID: emailVerificationOTPSentID,
			SentAt:                     formatEventTime(sentAt),
			Scope:                      EmailVerificationOTPGeneratedScope{EmailVerificationOTPGeneratedID: emailVerificationOTPGeneratedID},
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
			Scope:                    UserRegisteredScope{UserRegisteredID: userRegisteredID},
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
			Scope:             UserRegisteredScope{UserRegisteredID: userRegisteredID},
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
			Scope:                      UserRegisteredScope{UserRegisteredID: userRegisteredID},
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
