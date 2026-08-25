package models

import (
	"fmt"

	"seek/internal/features/_shared/sharedmodels"
)

func CompareServices(db, csv []*Service) []sharedmodels.Diff[Service] {
	keyFn := func(s *Service) string {
		return fmt.Sprintf("%s|%s|%d|%s", s.StudentID, s.ServiceName, s.FrequencyCount, s.FrequencyType)
	}
	diffFields := func(a, b *Service) []string {
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
		if a.LocationID != b.LocationID {
			changed = append(changed, "LocationID")
		}
		if a.ProviderID != b.ProviderID {
			changed = append(changed, "ProviderID")
		}
		if a.StartDate != b.StartDate {
			changed = append(changed, "StartDate")
		}
		if a.EndDate != b.EndDate {
			changed = append(changed, "EndDate")
		}
		return changed
	}

	return sharedmodels.Compare(db, csv, keyFn, diffFields)
}
