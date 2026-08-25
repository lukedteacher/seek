package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
	iepModels "seek/internal/features/ieps/models"
)

type StudentWithIEP struct {
	ID                  string                `json:"id"`
	MARSSID             string                `json:"marss_id"`
	sharedmodels.Person                       // embeds given, chosen, & family name, and email fields
	Grade               sharedmodels.Grade    `json:"grade"`
	HomeroomID          string                `json:"homeroom_id"`
	PlanType            sharedmodels.PlanType `json:"plan_type"`
	CaseManagerID       string                `json:"case_manager_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	IEP                 iepModels.IEP         `json:"iep"`
}
