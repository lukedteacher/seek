package models

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
