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
			Fields: []string{e.FieldScopeEducatorID},
			EventTypes: []string{
				e.EventEducatorCreated.String(),
				e.EventEducatorUpdated.String(),
				e.EventEducatorArchived.String(),
				e.EventEducatorDeleted.String(),
			},
		},
		{
			Name:   "period_scope",
			Fields: []string{p.FieldPeriodScopeID},
			EventTypes: []string{
				p.EventPeriodCreated.String(),
				p.EventPeriodUpdated.String(),
				p.EventPeriodArchived.String(),
				p.EventPeriodDeleted.String(),
			},
		},
		{
			Name:   "student_scope",
			Fields: []string{s.FieldScopeStudentID},
			EventTypes: []string{
				s.EventStudentCreated.String(),
				s.EventStudentUpdated.String(),
				s.EventStudentArchived.String(),
				s.EventStudentDeleted.String(),
			},
		},
	}
}
