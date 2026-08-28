package models

import (
	"fmt"

	"seek/internal/features/_shared/sharedmodels"
)

func CompareStudents(db, csv []Student) []sharedmodels.Diff[Student] {
	keyFn := func(s Student) string {
		return fmt.Sprintf("%s|%s|%s", s.Birthdate, s.GivenName, s.FamilyName)
	}
	diffFields := func(a, b Student) []string {
		changed := []string{}
		if a.MARSSID != b.MARSSID {
			changed = append(changed, "MARSSID")
		}
		if a.ChosenName != b.ChosenName {
			changed = append(changed, "ChosenName")
		}
		if a.Grade != b.Grade {
			changed = append(changed, "Grade")
		}
		if a.PlanType != b.PlanType {
			changed = append(changed, "PlanType")
		}
		return changed
	}

	return sharedmodels.Compare(db, csv, keyFn, diffFields)
}
