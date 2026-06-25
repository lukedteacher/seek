package eventcatalog

import (
	"seek/internal/eventstore"
	"seek/internal/features/period"
	"seek/internal/features/student"
	"seek/internal/features/teacher"
)

func BoundaryIndexes() []eventstore.BoundaryIndexDefinition {
	return []eventstore.BoundaryIndexDefinition{
		{
			Name:       "period_scope",
			Fields:     []string{period.PeriodScopeIDField},
			EventTypes: []string{period.PeriodCreated, period.PeriodUpdated, period.PeriodDeleted},
		},
		{
			Name:       "student_scope",
			Fields:     []string{student.StudentScopeIDField},
			EventTypes: []string{student.StudentCreated, student.StudentUpdated, student.StudentDeleted},
		},
		{
			Name:       "teacher_scope",
			Fields:     []string{teacher.TeacherScopeIDField},
			EventTypes: []string{teacher.TeacherCreated, teacher.TeacherUpdated, teacher.TeacherDeleted},
		},
	}
}
