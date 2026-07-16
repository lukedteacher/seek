package auth

import (
	"testing"
	"time"

	"seek/internal/protectedpii"
)

func TestPasswordResetCompletedEventMatchesFrasesShape(t *testing.T) {
	resetAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)

	event := NewPasswordResetCompletedEvent("completed-id", resetAt, "request-id", "user-id", "hash", nil)

	if event.EventID != "completed-id" {
		t.Fatalf("event id = %q, want completed-id", event.EventID)
	}
	if event.EventType != PasswordResetCompleted {
		t.Fatalf("event type = %q, want %q", event.EventType, PasswordResetCompleted)
	}
	if event.Data[PasswordResetCompletedIDField] != "completed-id" {
		t.Fatalf("password_reset_completed_id = %v, want completed-id", event.Data[PasswordResetCompletedIDField])
	}
	if event.Data[PasswordResetCompletedResetAtField] != resetAt.Format(time.RFC3339) {
		t.Fatalf("reset_at = %v, want %s", event.Data[PasswordResetCompletedResetAtField], resetAt.Format(time.RFC3339))
	}
	if event.Data[PasswordResetCompletedPasswordHashField] != "hash" {
		t.Fatalf("password_hash = %v, want hash", event.Data[PasswordResetCompletedPasswordHashField])
	}

	scope, ok := event.Data["scope"].(map[string]any)
	if !ok {
		t.Fatal("scope missing or wrong type")
	}
	if scope[PasswordResetRequestedIDField] != "request-id" {
		t.Fatalf("scope.password_reset_requested_id = %v, want request-id", scope[PasswordResetRequestedIDField])
	}
	if scope[UserRegisteredIDField] != "user-id" {
		t.Fatalf("scope.user_registered_id = %v, want user-id", scope[UserRegisteredIDField])
	}
}

func TestEventSpecificIDMatchesEventID(t *testing.T) {
	changedAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)

	passwordChanged := NewPasswordChangedEvent("password-changed-id", changedAt, "user-id", "hash", nil)
	if passwordChanged.EventID != passwordChanged.Data[PasswordChangedIDField] {
		t.Fatalf("password changed event id = %q, password_changed_id = %v", passwordChanged.EventID, passwordChanged.Data[PasswordChangedIDField])
	}

	subjectKey, err := protectedpii.GenerateSubjectDataKey()
	if err != nil {
		t.Fatal(err)
	}
	userNameChanged := NewUserNameChangedEvent("name-changed-id", "Ada Lovelace", changedAt, "user-id", subjectKey, nil)
	if userNameChanged.EventID != userNameChanged.Data[UserNameChangedIDField] {
		t.Fatalf("name changed event id = %q, user_name_changed_id = %v", userNameChanged.EventID, userNameChanged.Data[UserNameChangedIDField])
	}
}
