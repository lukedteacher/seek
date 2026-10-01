package sharedmodels

import (
	"encoding/json"
	"fmt"
)

type ServiceType string

const (
	ServiceTypeUnassigned           ServiceType = "unassigned"
	ServiceTypeExecutiveFunctioning ServiceType = "EF"
	ServiceTypeSEL                  ServiceType = "SEL"
	ServiceTypeReading              ServiceType = "reading"
	ServiceTypeWriting              ServiceType = "writing"
	ServiceTypeMath                 ServiceType = "math"
	ServiceTypeOccupationalTherapy  ServiceType = "OT"
	ServiceTypeSpeech               ServiceType = "speech"
	ServiceTypeSocialWork           ServiceType = "SW"
	ServiceTypeDAPE                 ServiceType = "DAPE"
	ServiceTypeNursing              ServiceType = "nursing"
	ServiceTypeVisionInstruction    ServiceType = "VI"
	ServiceTypeBehavioralSupport    ServiceType = "BS"
	ServiceTypeEmotionalRegulation  ServiceType = "ER"
	ServiceTypeFunctionalSkills     ServiceType = "FS"
)

// knownServiceTypes is used for validation in UnmarshalJSON
var knownServiceTypes = []ServiceType{
	ServiceTypeUnassigned,
	ServiceTypeExecutiveFunctioning,
	ServiceTypeSEL,
	ServiceTypeReading,
	ServiceTypeWriting,
	ServiceTypeMath,
	ServiceTypeOccupationalTherapy,
	ServiceTypeSpeech,
	ServiceTypeSocialWork,
	ServiceTypeDAPE,
	ServiceTypeNursing,
	ServiceTypeVisionInstruction,
	ServiceTypeBehavioralSupport,
	ServiceTypeEmotionalRegulation,
	ServiceTypeFunctionalSkills,
}

var ServiceTypeList = []ServiceType{
	ServiceTypeExecutiveFunctioning,
	ServiceTypeSEL,
	ServiceTypeReading,
	ServiceTypeWriting,
	ServiceTypeMath,
	ServiceTypeOccupationalTherapy,
	ServiceTypeSpeech,
	ServiceTypeSocialWork,
	ServiceTypeDAPE,
	ServiceTypeNursing,
	ServiceTypeVisionInstruction,
	ServiceTypeBehavioralSupport,
	ServiceTypeEmotionalRegulation,
	ServiceTypeFunctionalSkills,
}

// returns a simple string of the service type
func (st ServiceType) String() string { return string(st) }

// returns the short form (abbreviation) of the service type
func (st ServiceType) ShortString() string {
	return map[ServiceType]string{
		ServiceTypeUnassigned:           "unassigned",
		ServiceTypeExecutiveFunctioning: "EF",
		ServiceTypeSEL:                  "SEL",
		ServiceTypeReading:              "reading",
		ServiceTypeWriting:              "writing",
		ServiceTypeMath:                 "math",
		ServiceTypeOccupationalTherapy:  "OT",
		ServiceTypeSpeech:               "speech",
		ServiceTypeSocialWork:           "SW",
		ServiceTypeDAPE:                 "DAPE",
		ServiceTypeNursing:              "nursing",
		ServiceTypeVisionInstruction:    "VI",
		ServiceTypeBehavioralSupport:    "BS",
		ServiceTypeEmotionalRegulation:  "ER",
		ServiceTypeFunctionalSkills:     "FS",
	}[st]
}

// returns the full descriptive name of the service type
func (st ServiceType) LongString() string {
	return map[ServiceType]string{
		ServiceTypeUnassigned:           "unassigned",
		ServiceTypeExecutiveFunctioning: "executive functioning",
		ServiceTypeSEL:                  "social emotional learning",
		ServiceTypeReading:              "reading",
		ServiceTypeWriting:              "writing",
		ServiceTypeMath:                 "mathematics",
		ServiceTypeOccupationalTherapy:  "occupational therapy",
		ServiceTypeSpeech:               "speech therapy",
		ServiceTypeSocialWork:           "social work",
		ServiceTypeDAPE:                 "developmentallly appropriate physical education",
		ServiceTypeNursing:              "nursing",
		ServiceTypeVisionInstruction:    "vision instruction",
		ServiceTypeBehavioralSupport:    "behavioral support",
		ServiceTypeEmotionalRegulation:  "emotional regulation",
		ServiceTypeFunctionalSkills:     "functional skills",
	}[st]
}

// returns the icon name for each service type
// use: @icon.Icon(service.ServiceType.IconName())(icon.Props{Size: "16"})
func (st ServiceType) IconName() string {
	if name, ok := map[ServiceType]string{
		ServiceTypeUnassigned:           "circle-question-mark",
		ServiceTypeExecutiveFunctioning: "brain-cog",
		ServiceTypeSEL:                  "heart-handshake",
		ServiceTypeReading:              "book-open",
		ServiceTypeWriting:              "notebook-pen",
		ServiceTypeMath:                 "calculator",
		ServiceTypeOccupationalTherapy:  "pencil",
		ServiceTypeSpeech:               "speech",
		ServiceTypeSocialWork:           "message-circle-heart",
		ServiceTypeDAPE:                 "biceps-flexed",
		ServiceTypeNursing:              "stethoscope",
		ServiceTypeVisionInstruction:    "eye",
		ServiceTypeBehavioralSupport:    "plus",
		ServiceTypeEmotionalRegulation:  "plus",
		ServiceTypeFunctionalSkills:     "plus",
	}[st]; ok {
		return name
	}
	return "help-circle" // fallback for unknown/custom types
}

func ParseServiceType(s string) (ServiceType, error) {
	if s == "" {
		return ServiceTypeUnassigned, nil
	}
	for _, valid := range knownServiceTypes {
		if s == string(valid) {
			return ServiceType(s), nil
		}
	}
	return "", fmt.Errorf("invalid ServiceType: %q", s)
}

func (s ServiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *ServiceType) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if str == "" {
		*s = ServiceTypeUnassigned
		return nil
	}
	for _, valid := range knownServiceTypes {
		if str == string(valid) {
			*s = ServiceType(str)
			return nil
		}
	}
	return fmt.Errorf("invalid ServiceType: %q", str)
}
