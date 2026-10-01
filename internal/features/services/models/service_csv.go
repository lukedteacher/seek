package models

import (
	"fmt"
	"seek/internal/features/_shared/sharedmodels"
	"strings"
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
	c.ServiceName = strings.Trim(c.ServiceName, " ")
	serviceType, ok := serviceNameTypeMap[c.ServiceName]
	if !ok {
		serviceType = sharedmodels.ServiceTypeUnassigned
	}
	return Service{
		ID:              c.ID,
		IEPID:           c.IEPID,
		StudentID:       c.StudentID,
		StudentMARSSID:  c.StudentMARSSID,
		ServiceName:     c.ServiceName,
		ServiceType:     serviceType,
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

var serviceNameTypeMap = map[string]sharedmodels.ServiceType{
	"School Social Work Services":                        sharedmodels.ServiceTypeSocialWork,
	"Occupational Therapy":                               sharedmodels.ServiceTypeOccupationalTherapy,
	"Specialized Instruction: Social Skills Training":    sharedmodels.ServiceTypeSEL,
	"Specialized Instruction: Mathematics":               sharedmodels.ServiceTypeMath,
	"Specialized Instruction: Executive Functioning":     sharedmodels.ServiceTypeExecutiveFunctioning,
	"Specialized Instruction: Reading":                   sharedmodels.ServiceTypeReading,
	"Written Expression":                                 sharedmodels.ServiceTypeWriting,
	"Push-In Support: Social-Emotional Learning":         sharedmodels.ServiceTypeSEL,
	"Specialized Instruction: Social-Emotional Learning": sharedmodels.ServiceTypeSEL,
	"Specialized Instruction: Written Language":          sharedmodels.ServiceTypeWriting,
	"Social Skills":                                      sharedmodels.ServiceTypeSEL,
	"Speech/Language: Articulation":                      sharedmodels.ServiceTypeSpeech,
	"Social Emotional Learning":                          sharedmodels.ServiceTypeSEL,
	"Specialized Instruction: Math":                      sharedmodels.ServiceTypeMath,
	"Speech/Language: Language Instruction":              sharedmodels.ServiceTypeSpeech,
	"Developmental Adaptive Physical Education":          sharedmodels.ServiceTypeDAPE,
	"Nursing Services":                                   sharedmodels.ServiceTypeNursing,
	"Social-Emotional Learning: Pull-Out Small Group":    sharedmodels.ServiceTypeSEL,
	"Social-Emotional Learning: Push-In Support":         sharedmodels.ServiceTypeSEL,
	"Vision Instruction":                                 sharedmodels.ServiceTypeVisionInstruction,
	"Reading: Pull-Out Instruction":                      sharedmodels.ServiceTypeReading,
	"Writing: Pull-Out Instruction":                      sharedmodels.ServiceTypeWriting,
	"Mathematics: Push-In Support":                       sharedmodels.ServiceTypeMath,
	"Occupational Therapy: Indirect Support":             sharedmodels.ServiceTypeOccupationalTherapy,
	"Behavioral Support":                                 sharedmodels.ServiceTypeBehavioralSupport,
	"Emotional Regulation":                               sharedmodels.ServiceTypeEmotionalRegulation,
	"Functional Skills":                                  sharedmodels.ServiceTypeFunctionalSkills,
	"Specialized Instruction: Emotional Regulation":      sharedmodels.ServiceTypeEmotionalRegulation,
	"Emotional Regulation: Social Work":                  sharedmodels.ServiceTypeSocialWork,
	"Speech/Language: Language and Articulation":         sharedmodels.ServiceTypeSpeech,
	"Academic: Writing":                                  sharedmodels.ServiceTypeWriting,
	"Academic: Math":                                     sharedmodels.ServiceTypeMath,
	"Academic: Reading":                                  sharedmodels.ServiceTypeReading,
	"Academics: Writing":                                 sharedmodels.ServiceTypeWriting,
	"Academics: Math":                                    sharedmodels.ServiceTypeMath,
	"Academics: Reading":                                 sharedmodels.ServiceTypeReading,
	"Speech/Language: Articulation and Language":         sharedmodels.ServiceTypeSpeech,
	"Speech/Language - Articulation and Phonology":       sharedmodels.ServiceTypeSpeech,
}
