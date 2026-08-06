package eventcatalog

import (
	"seek/internal/eventstore"
	e "seek/internal/features/educators/events"
	p "seek/internal/features/periods/events"
	s "seek/internal/features/students/events"
)

func BoundaryIndexes() []eventstore.BoundaryIndexDefinition {
	return []eventstore.BoundaryIndexDefinition{
		{
			Name:   "educator_scope",
			Fields: []string{e.FieldEducatorScopeID},
			EventTypes: []string{
				e.EducatorCreated,
				e.EducatorUpdated,
				e.EducatorArchived,
				e.EducatorDeleted,
			},
		},
		{
			Name:   "period_scope",
			Fields: []string{p.FieldPeriodScopeID},
			EventTypes: []string{
				p.PeriodCreated,
				p.PeriodUpdated,
				p.PeriodArchived,
				p.PeriodDeleted,
			},
		},
		{
			Name:   "student_scope",
			Fields: []string{s.FieldStudentScopeID},
			EventTypes: []string{
				s.StudentCreated,
				s.StudentUpdated,
				s.StudentArchived,
				s.StudentDeleted,
			},
		},
	}
}
