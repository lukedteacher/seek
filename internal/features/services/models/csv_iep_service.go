package models

import (
	"fmt"
	"seek/internal/features/_shared/sharedmodels"
	"time"
)

type CSVService struct {
	ID              string                   `json:"id"`
	IEPID           string                   `json:"iep_id"`
	StudentID       string                   `json:"student_id"`
	StudentMARSSID  string                   `json:"student_marss_id" csv:"MARSS ID"`
	ServiceName     string                   `json:"service_name" csv:"Service"`
	ServiceType     sharedmodels.ServiceType `json:"service_type"`
	IndirectMinutes int                      `json:"indirect_minutes" csv:"Indirect minutes"`
	DirectMinutes   int                      `json:"direct_minutes" csv:"Direct minutes"`
	FrequencyCount  int                      `json:"frequency_count" csv:"Frequency count"`
	FrequencyType   string                   `json:"frequency_type" csv:"Frequency"`
	Location        string                   `json:"location"`
	StartDate       CSVTime                  `json:"start_date" csv:"Start date"`
	EndDate         CSVTime                  `json:"end_date" csv:"End date"`
	Provider        string                   `json:"provider" csv:"Provider"`
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

func NewCSVService() *Service {
	return &Service{}
}

func NewModelFromCSVRow(c CSVService) Service {
	return Service{
		ID:              c.ID,
		IEPID:           c.IEPID,
		StudentID:       c.StudentID,
		StudentMARSSID:  c.StudentMARSSID,
		ServiceName:     c.ServiceName,
		ServiceType:     c.ServiceType,
		IndirectMinutes: c.IndirectMinutes,
		DirectMinutes:   c.DirectMinutes,
		FrequencyCount:  c.FrequencyCount,
		FrequencyType:   c.FrequencyType,
		LocationID:      c.Location,
		StartDate:       sharedmodels.DateOnly(c.StartDate),
		EndDate:         sharedmodels.DateOnly(c.EndDate),
		Provider:        c.Provider,
	}
}

func CompareServices(db, csv []Service) []sharedmodels.Diff[Service] {
	keyFn := func(s Service) string {
		return fmt.Sprintf("%s|%s|%d|%s", s.StudentID, s.ServiceName, s.FrequencyCount, s.FrequencyType)
	}
	diffFields := func(a, b Service) []string {
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
