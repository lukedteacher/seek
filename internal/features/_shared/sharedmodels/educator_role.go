package sharedmodels

import "encoding/json"

type EducatorRole string

const (
	EducatorRoleNone                 EducatorRole = "na"
	EducatorRoleGenEdTeacher         EducatorRole = "get"
	EducatorRoleCoTeacher            EducatorRole = "cot"
	EducatorRoleResourceRoomTeacher  EducatorRole = "rrt"
	EducatorRoleCaseManager          EducatorRole = "cmr"
	EducatorRoleServiceProvider      EducatorRole = "spr"
	EducatorRoleEducationalAssistant EducatorRole = "eda"
	EducatorRoleAdmin                EducatorRole = "adm"
)

var EducatorRoleList = []EducatorRole{
	EducatorRoleNone,
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
	long := map[EducatorRole]string{
		EducatorRoleNone:                 "none",
		EducatorRoleGenEdTeacher:         "general education teacher",
		EducatorRoleCoTeacher:            "co-teacher",
		EducatorRoleResourceRoomTeacher:  "resource room teacher",
		EducatorRoleCaseManager:          "case manager",
		EducatorRoleServiceProvider:      "service provider",
		EducatorRoleEducationalAssistant: "educational assistant",
		EducatorRoleAdmin:                "admin",
	}
	if s, ok := long[er]; ok {
		return s
	}
	return string(er)
}

func (r *EducatorRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	for _, role := range EducatorRoleList {
		if string(role) == s || role.StringLong() == s {
			*r = role
			return nil
		}
	}
	*r = EducatorRole(s)
	return nil
}
