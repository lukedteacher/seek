package auth

import (
	"testing"
	"time"
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
	if event.Data[PasswordResetCompletedEventID] != "completed-id" {
		t.Fatalf("password_reset_completed_id = %v, want completed-id", event.Data[PasswordResetCompletedEventID])
	}
	if event.Data[FieldPasswordResetCompletedResetAt] != resetAt.Format(time.RFC3339) {
		t.Fatalf("reset_at = %v, want %s", event.Data[FieldPasswordResetCompletedResetAt], resetAt.Format(time.RFC3339))
	}
	if event.Data[FieldPasswordResetCompletedPasswordHash] != "hash" {
		t.Fatalf("password_hash = %v, want hash", event.Data[FieldPasswordResetCompletedPasswordHash])
	}

	scope, ok := event.Data["scope"].(map[string]any)
	if !ok {
		t.Fatal("scope missing or wrong type")
	}
	if scope[PasswordResetRequestedEventID] != "request-id" {
		t.Fatalf("scope.password_reset_requested_id = %v, want request-id", scope[PasswordResetRequestedEventID])
	}
	if scope[UserRegisteredEventID] != "user-id" {
		t.Fatalf("scope.user_registered_id = %v, want user-id", scope[UserRegisteredEventID])
	}
}

func TestEventSpecificIDMatchesEventID(t *testing.T) {
	changedAt := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)

	passwordChanged := NewPasswordChangedEvent("password-changed-id", changedAt, "user-id", "hash", nil)
	if passwordChanged.EventID != passwordChanged.Data[PasswordChangedEventID] {
		t.Fatalf("password changed event id = %q, password_changed_id = %v", passwordChanged.EventID, passwordChanged.Data[PasswordChangedEventID])
	}
}
