package models

import (
	"seek/internal/features/_shared/models"
)

type Educator struct {
	models.Person        // embeds given, chosen, & family name fields
	ID            string `json:"id" display:"ID"`
	Email         string `json:"email" display:"email"`
	Role          string `json:"role" display:"role"`
}

func NewEducator() *Educator {
	return &Educator{}
}
