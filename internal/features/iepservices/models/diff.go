package models

import "seek/internal/features/_shared/sharedmodels"

func CompareIEPServices(db, csv []*IEPService) []sharedmodels.Diff[IEPService] {
	keyFn := func(s *IEPService) string {
		return s.StudentID + "|" + s.ServiceName
	}
	diffFields := func(a, b *IEPService) []string {
		changed := []string{}
		if a.ServiceName != b.ServiceName {
			changed = append(changed, "ServiceName")
		}
		if a.ServiceType != b.ServiceType {
			changed = append(changed, "ServiceType")
		}
		if a.IndirectMinutes != b.IndirectMinutes {
			changed = append(changed, "IndirectMinutes")
		}
		if a.DirectMinutes != b.DirectMinutes {
			changed = append(changed, "DirectMinutes")
		}
		if a.FrequencyCount != b.FrequencyCount {
			changed = append(changed, "FrequencyCount")
		}
		if a.FrequencyType != b.FrequencyType {
			changed = append(changed, "FrequencyType")
		}
		if a.Location != b.Location {
			changed = append(changed, "Location")
		}
		if a.Provider != b.Provider {
			changed = append(changed, "Provider")
		}
		if a.StartDate != b.StartDate {
			changed = append(changed, "StartDate")
		}
		if a.EndDate != b.EndDate {
			changed = append(changed, "EndDate")
		}
		// skip ID, CreatedAt, UpdatedAt, ArchivedAt
		return changed
	}

	return sharedmodels.Compare(db, csv, keyFn, diffFields)
}
