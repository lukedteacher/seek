package models

type Profile struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Bio  string `json:"bio"`
}
