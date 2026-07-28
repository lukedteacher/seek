package models

import (
	"seek/internal/features/_shared/sharedmodels"
)

type Student struct {
	ID                  string             `json:"id"`
	sharedmodels.Person                    // embeds given, chosen, & family name, and email fields
	Grade               sharedmodels.Grade `json:"grade"`
	Homeroom            string             `json:"homeroom"`
	CaseManager         string             `json:"case_manager"`
}

func NewStudent() *Student {
	return &Student{}
}
