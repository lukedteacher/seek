package eventcatalog

import (
	"seek/internal/eventstore"
	period "seek/internal/features/periods/events"
	schedule "seek/internal/features/schedules/events"
	student "seek/internal/features/students/events"
	teacher "seek/internal/features/teachers/events"
)

func BoundaryIndexes() []eventstore.BoundaryIndexDefinition {
	return []eventstore.BoundaryIndexDefinition{
		{
			Name:       "period_scope",
			Fields:     []string{period.PeriodScopeIDField},
			EventTypes: []string{period.PeriodCreated, period.PeriodUpdated, period.PeriodDeleted},
		},
		{
			Name:       "schedule_scope",
			Fields:     []string{schedule.ScheduleScopeIDField},
			EventTypes: []string{schedule.ScheduleCreated, schedule.ScheduleUpdated, schedule.ScheduleDeleted},
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
