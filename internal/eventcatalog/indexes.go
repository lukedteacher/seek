package eventcatalog

import (
	"seek/internal/eventstore"
	"seek/internal/features/student"
)

func BoundaryIndexes() []eventstore.BoundaryIndexDefinition {
	return []eventstore.BoundaryIndexDefinition{
		{
			Name:       "student_scope",
			Fields:     []string{student.StudentScopeIDField},
			EventTypes: []string{student.StudentCreated, student.StudentRenamed, student.StudentCompleted, student.StudentReopened, student.StudentDeleted},
		},
	}
}
