package models

type Profile struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}
