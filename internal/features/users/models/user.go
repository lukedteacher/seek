package models

import "context"

type contextKey string

var UserKey contextKey = "user"

func GetUserFromContext(ctx context.Context) User {
	if user, ok := ctx.Value(UserKey).(User); ok {
		return user
	}
	return User{}
}

type User struct {
	ID               string
	UserRegisteredID string
	Email            string
	EmailVerified    bool
	Role             string
	Image            string
	Bio              string
	HeaderImageURL   string
}
