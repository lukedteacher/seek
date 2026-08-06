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
