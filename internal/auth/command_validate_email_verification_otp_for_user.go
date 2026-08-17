package auth

import (
	"context"

	"seek/internal/features/users/models"
)

type AuthUserByIDReader interface {
	UserByIDOrRegisteredID(ctx context.Context, id string) (models.User, error)
}
