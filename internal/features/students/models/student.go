package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type Student struct {
	ID                  string                `json:"id"`
	MARSSID             string                `json:"marss_id" csv:"MARSS ID"`
	sharedmodels.Person                       // embeds given, chosen, & family name, and email fields
	Grade               sharedmodels.Grade    `json:"grade" csv:"grade"`
	HomeroomID          string                `json:"homeroom_id"`
	PlanType            sharedmodels.PlanType `json:"plan_type"`
	CaseManagerID       string                `json:"case_manager_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

func NewStudent() *Student {
	return &Student{}
}
