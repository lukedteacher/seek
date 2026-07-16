package auth

import (
	"context"

	"seek/internal/email"
)

type EmailSender interface {
	Send(ctx context.Context, message email.Message) error
}
