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

// read model writer functions

func (m *ReadModel) AddStudentBookmark(ctx context.Context, event StudentBookmarkAddedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddStudentBookmark(conn, dbsql.AddStudentBookmarkParams{
			UserId:                   event.UserID,
			StudentId:                event.StudentID,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			AddedAt:                  event.AddedAt,
		})
	})
}

func (m *ReadModel) RemoveStudentBookmark(ctx context.Context, event StudentBookmarkRemovedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveStudentBookmark(conn, dbsql.RemoveStudentBookmarkParams{
			UserId:    event.UserID,
			StudentId: event.StudentID,
		})
	})
}
