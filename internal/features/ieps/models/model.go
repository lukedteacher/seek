package models

import (
	"time"

	"seek/internal/features/_shared/sharedmodels"
)

type IEP struct {
	ID                     string                `json:"id"`
	StudentID              string                `json:"student_id"`
	StudentMARSSID         string                `json:"student_marss_id"`
	PlanManagerSPEDFormsID string                `json:"teammember_id"`
	Disability1            DisabilityCode        `json:"disability_1"`
	Disability2            DisabilityCode        `json:"disability_2"`
	FederalSetting         int                   `json:"federal_setting"`
	MeetingDate            sharedmodels.DateOnly `json:"meeting_date"`
	IEPDueDate             sharedmodels.DateOnly `json:"iep_due_date"`
	LastEvalDate           sharedmodels.DateOnly `json:"last_eval_date"`
	EvalDueDate            sharedmodels.DateOnly `json:"eval_due_date"`
	AmendedDate            sharedmodels.DateOnly `json:"amended_date"`
	IEPType                IEPType               `json:"iep_type"`
	SpecialTransportation  bool                  `json:"special_transportation"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

func NewIEP() *IEP {
	return &IEP{}
}
