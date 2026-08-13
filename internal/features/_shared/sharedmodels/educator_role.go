package sharedmodels

type EducatorRole string

const (
	EducatorRoleCaseManager          EducatorRole = "case manager"
	EducatorRoleRResRoomTeacher      EducatorRole = "resource room teacher"
	EducatorRoleGenEdTeacher         EducatorRole = "general education teacher"
	EducatorRoleServiceProvider      EducatorRole = "service provider"
	EducatorRoleCoTeacher            EducatorRole = "co-teacher"
	EducatorRoleEducationalAssistant EducatorRole = "educational assistant"
	EducatorRoleAdmin                EducatorRole = "admin"
)

var EducatorRoleList = []EducatorRole{
	EducatorRoleCaseManager,
	EducatorRoleCoTeacher,
	EducatorRoleEducationalAssistant,
	EducatorRoleGenEdTeacher,
	EducatorRoleRResRoomTeacher,
	EducatorRoleServiceProvider,
	EducatorRoleAdmin,
}

func (er EducatorRole) String() string {
	return string(er)
}

func (er EducatorRole) ShortString() string {
	short := map[EducatorRole]string{
		EducatorRoleCaseManager:          "CM",
		EducatorRoleRResRoomTeacher:      "RR",
		EducatorRoleGenEdTeacher:         "GenEd",
		EducatorRoleServiceProvider:      "SP",
		EducatorRoleCoTeacher:            "CoT",
		EducatorRoleEducationalAssistant: "EA",
		EducatorRoleAdmin:                "admin",
	}
	if s, ok := short[er]; ok {
		return s
	}
	return string(er)
}
