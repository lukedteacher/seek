package events

import (
	"context"

	"seek/internal/appdb"
	"seek/internal/dbsql"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

type CaseManagerStudentReadModelWriter interface {
	AddStudentToCaseload(ctx context.Context, event StudentAddedToCaseloadProjection) error
	RemoveStudentFromCaseload(ctx context.Context, event StudentRemovedFromCaseloadProjection) error
}

// period educator read model writer functions

func (m *ReadModel) AddStudentToCaseload(ctx context.Context, event StudentAddedToCaseloadProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddStudentToCaseload(conn, dbsql.AddStudentToCaseloadParams{
			EducatorId:               event.EducatorID,
			StudentId:                event.StudentID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemoveStudentFromCaseload(ctx context.Context, event StudentRemovedFromCaseloadProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveStudentFromCaseload(conn, dbsql.RemoveStudentFromCaseloadParams{
			EducatorId: event.EducatorID,
			StudentId:  event.StudentID,
		})
	})
}
