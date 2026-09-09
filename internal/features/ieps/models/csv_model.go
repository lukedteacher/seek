package models

import (
	"fmt"
	"seek/internal/features/_shared/sharedmodels"
	"strconv"
	"time"
)

type IEPCSVRow struct {
	StudentID              string  `json:"student_id"`
	StudentMARSSID         string  `json:"student_marss_id" csv:"student_id"`
	PlanManagerSPEDFormsID string  `json:"plan_manager_sped_forms_id" csv:"teammember_id"`
	Disability1            string  `json:"disability_1" csv:"disability1"`
	Disability2            string  `json:"disability_2" csv:"disability1"`
	FederalSetting         int     `json:"federal_setting" csv:"federalsetting"`
	MeetingDate            CSVTime `json:"meeting_date" csv:"meeting_date"`
	IEPDueDate             CSVTime `json:"iep_due_date" csv:"iep_due"`
	LastEvalDate           CSVTime `json:"last_eval_date" csv:"lastcompevaldate"`
	EvalDueDate            CSVTime `json:"eval_due_date" csv:"eval_due"`
	AmendedDate            CSVTime `json:"amended_date"`
	InitialIEP             bool    `json:"initial_iep" csv:"initialiep"`
	AnnualIEP              bool    `json:"annual_iep" csv:"annualiep"`
	InterimIEP             bool    `json:"interim_iep" csv:"interimiep"`
	SpecialTransportation  bool    `json:"special_transportation"`
}

type CSVTime time.Time

func (ct *CSVTime) UnmarshalCSV(csv string) error {
	t, err := time.Parse("01/02/2006 03:04 PM", csv)
	if err != nil {
		return err
	}
	*ct = CSVTime(t)
	return nil
}

func NewModelFromCSVRow(r IEPCSVRow) IEP {
	disability1, _ := strconv.Atoi(r.Disability1)
	disability2, _ := strconv.Atoi(r.Disability2)
	return IEP{
		StudentID:              r.StudentID,
		StudentMARSSID:         r.StudentMARSSID,
		PlanManagerSPEDFormsID: r.PlanManagerSPEDFormsID,
		Disability1:            DisabilityCode(disability1),
		Disability2:            DisabilityCode(disability2),
		FederalSetting:         r.FederalSetting,
		MeetingDate:            sharedmodels.DateOnly(r.MeetingDate),
		IEPDueDate:             sharedmodels.DateOnly(r.IEPDueDate),
		LastEvalDate:           sharedmodels.DateOnly(r.EvalDueDate),
		EvalDueDate:            sharedmodels.DateOnly(r.EvalDueDate),
		AmendedDate:            sharedmodels.DateOnly(r.AmendedDate),
		IEPType:                iepTypeFromBools(r.InitialIEP, r.AnnualIEP, r.InterimIEP),
		SpecialTransportation:  r.SpecialTransportation,
	}
}

func CompareServices(db, csv []IEP) []sharedmodels.Diff[IEP] {
	keyFn := func(i IEP) string {
		return fmt.Sprintf("%s|%s", i.StudentID, i.MeetingDate)
	}
	diffFields := func(a, b IEP) []string {
		changed := []string{}
		if a.MeetingDate != b.MeetingDate {
			changed = append(changed, "MeetingDate")
		}
		return changed
	}

	return sharedmodels.Compare(db, csv, keyFn, diffFields)
}

func iepTypeFromBools(initial, annual, interim bool) IEPType {
	if initial {
		return IEPTypeInitial
	}
	if annual {
		return IEPTypeAnnual // IEPTypeAnnual
	}
	if interim {
		return IEPTypeInterim // IEPTypeInterim
	}
	return IEPTypeUnknown
}
