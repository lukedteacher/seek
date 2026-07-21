package models

import "strings"

type User struct {
	ID               string
	UserRegisteredID string
	Name             string
	Username         string
	Email            string
	EmailVerified    bool
	Role             string
	Image            string
	Bio              string
	HeaderImageURL   string
}

func DisplayName(user User) string {
	if user.Name != "" {
		return user.Name
	}
	if user.Username != "" {
		return user.Username
	}
	return user.Email
}

func UserInitials(user User) string {
	name := strings.TrimSpace(DisplayName(user))
	if name == "" {
		return "?"
	}

	parts := strings.Fields(name)
	if len(parts) == 1 {
		return strings.ToUpper(string([]rune(parts[0])[0]))
	}

	first := []rune(parts[0])
	last := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(first[0]) + string(last[0]))
}
