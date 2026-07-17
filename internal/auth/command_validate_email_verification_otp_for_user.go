package auth

import (
	"context"

	"seek/internal/commandlimits"
	"seek/internal/domain/models"
	"seek/internal/eventstore"
)

type AuthUserByIDReader interface {
	UserByIDOrRegisteredID(ctx context.Context, id string) (models.User, error)
}

func ValidateEmailVerificationOTPForUserCommandHandler(
	ctx context.Context,
	userID,
	code string,
	metadata eventstore.CommandMetadata,
	users AuthUserByIDReader,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
) error {
	if err := commandlimits.Assert(struct {
		UserID string
		Code   string
	}{UserID: userID, Code: code}); err != nil {
		return err
	}
	println(userID)
	user, err := users.UserByIDOrRegisteredID(ctx, userID)
	if err != nil {
		println("is this the error?")
		return err
	}
	return ValidateEmailVerificationOTPCommandHandler(ctx, ValidateEmailVerificationOTPCommand{
		User:     user,
		Code:     code,
		Metadata: metadata,
	}, saver, retriever)
}
