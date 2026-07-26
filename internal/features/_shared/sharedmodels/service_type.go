package sharedmodels

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
)

var ServiceTypeList = []ServiceType{
	ServiceTypeExecutiveFunctioning,
	ServiceTypeSEL,
	ServiceTypeReading,
	ServiceTypeWriting,
	ServiceTypeMath,
	ServiceTypeOccupationalTherapy,
	ServiceTypeSpeech,
	ServiceTypeSocialWork,
}

// String returns the short form (abbreviation) of the service type.
func (st ServiceType) String() string { return string(st) }

// String returns the short form (abbreviation) of the service type.
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
	}[st]
}

// LongString returns the full descriptive name.
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
	}[st]
}

// returns the icon name for each service type
// use: @icon.Icon(service.ServiceType.IconName())(icon.Props{Size: "16"})
func (st ServiceType) IconName() string {
	if name, ok := map[ServiceType]string{
		ServiceTypeUnassigned:           "help-circle",
		ServiceTypeExecutiveFunctioning: "brain",
		ServiceTypeSEL:                  "heart-handshake",
		ServiceTypeReading:              "book-open",
		ServiceTypeWriting:              "pen-tool",
		ServiceTypeMath:                 "sigma",
		ServiceTypeOccupationalTherapy:  "hand",
		ServiceTypeSpeech:               "messages-square",
		ServiceTypeSocialWork:           "users",
	}[st]; ok {
		return name
	}
	return "help-circle" // fallback for unknown/custom types
}
