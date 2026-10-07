package sharedmodels

type EducatorRole string

const (
	EducatorRoleGenEdTeacher         EducatorRole = "get"
	EducatorRoleCoTeacher            EducatorRole = "cot"
	EducatorRoleResourceRoomTeacher  EducatorRole = "rrt"
	EducatorRoleCaseManager          EducatorRole = "cmr"
	EducatorRoleServiceProvider      EducatorRole = "spr"
	EducatorRoleEducationalAssistant EducatorRole = "eda"
	EducatorRoleAdmin                EducatorRole = "adm"
)

var EducatorRoleList = []EducatorRole{
	EducatorRoleGenEdTeacher,
	EducatorRoleCoTeacher,
	EducatorRoleResourceRoomTeacher,
	EducatorRoleCaseManager,
	EducatorRoleServiceProvider,
	EducatorRoleEducationalAssistant,
	EducatorRoleAdmin,
}

func (er EducatorRole) String() string {
	return string(er)
}

func (er EducatorRole) StringLong() string {
	short := map[EducatorRole]string{
		EducatorRoleGenEdTeacher:         "general education teacher",
		EducatorRoleCoTeacher:            "co-teacher",
		EducatorRoleResourceRoomTeacher:  "resource room teacher",
		EducatorRoleCaseManager:          "case manager",
		EducatorRoleServiceProvider:      "service provider",
		EducatorRoleEducationalAssistant: "educational assistant",
		EducatorRoleAdmin:                "admin",
	}
	if s, ok := short[er]; ok {
		return s
	}
	return string(er)
}
