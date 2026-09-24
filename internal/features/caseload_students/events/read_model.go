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
	println("add")
	println("sid", event.StudentID)
	println("eid", event.EducatorID)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		if err := dbsql.OnceAddStudentToCaseload(conn, dbsql.AddStudentToCaseloadParams{
			EducatorId:               event.EducatorID,
			StudentId:                event.StudentID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		}); err != nil {
			return err
		}
		return dbsql.OnceSetStudentCaseManager(conn, dbsql.SetStudentCaseManagerParams{
			CaseManagerId:            event.EducatorID,
			Id:                       event.StudentID,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.AddedAt),
		})
	})
}

func (m *ReadModel) RemoveStudentFromCaseload(ctx context.Context, event StudentRemovedFromCaseloadProjection) error {
	println("remove")
	println("sid", event.StudentID)
	println("eid", event.EducatorID)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		if err := dbsql.OnceRemoveStudentFromCaseload(conn, dbsql.RemoveStudentFromCaseloadParams{
			EducatorId: event.EducatorID,
			StudentId:  event.StudentID,
		}); err != nil {
			return err
		}
		return dbsql.OnceClearStudentCaseManager(conn, dbsql.ClearStudentCaseManagerParams{
			Id:                       event.StudentID,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.RemovedAt),
		})
	})
}
