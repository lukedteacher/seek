package models

import (
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/iepservices/models"
	"time"
)

type Student struct {
	ID                  string              `json:"id"`
	MARSSID             string              `json:"marss_id"`
	sharedmodels.Person                     // embeds given, chosen, & family name, and email fields
	Grade               sharedmodels.Grade  `json:"grade"`
	Homeroom            string              `json:"homeroom"`
	CaseManager         string              `json:"case_manager"`
	Services            []models.IEPService `json:"iep_services"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}

func NewStudent() *Student {
	return &Student{}
}
