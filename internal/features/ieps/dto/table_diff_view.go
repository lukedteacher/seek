package dto

import (
	"fmt"

	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
)

func CompareIEPs(db, csv []IEPView) []sharedmodels.Diff[IEPView] {
	keyFn := func(i IEPView) string {
		return fmt.Sprintf("%s|%s", i.StudentID, i.MeetingDate)
	}
	diffFields := func(a, b IEPView) []string {
		changed := []string{}
		if a.MeetingDate != b.MeetingDate {
			changed = append(changed, "MeetingDate")
		}
		return changed
	}

	return sharedmodels.Compare(db, csv, keyFn, diffFields)
}

func NewServiceDiffTableView(diffs []sharedmodels.Diff[IEPView]) shareddto.DiffTableView {
	return shareddto.NewDiffTableView(diffs, IEPTableConfig)
}
