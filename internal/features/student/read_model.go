package student

import (
	"context"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/views"
	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) List(ctx context.Context) ([]views.Student, error) {
	var rows []dbsql.ListStudentsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudents(conn)
		return err
	}); err != nil {
		return nil, err
	}
	students := make([]views.Student, 0, len(rows))
	for _, row := range rows {
		students = append(students, views.Student{
			ID:						row.ID,
			FirstName:		row.FirstName,
			ChosenName:		row.ChosenName,
			LastName:			row.LastName,
			Grade:				row.Grade,
			Homeroom:			row.Homeroom,
			CaseManager:	row.CaseManager,
			CreatedAt:		parseDBTime(row.CreatedAt),
			UpdatedAt:		parseDBTime(row.UpdatedAt),
		})
	}
	return students, nil
}

func (m *ReadModel) InsertCreatedStudent(ctx context.Context, event StudentCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceInsertCreatedStudent(conn, dbsql.InsertCreatedStudentParams{
			ID:                   event.ID,
			FirstName:                    event.FirstName,
			ChosenName:                    event.ChosenName,
			LastName:                    event.LastName,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) RenameStudent(ctx context.Context, event StudentRenamedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRenameStudent(conn, dbsql.RenameStudentParams{
			FirstName:                    event.FirstName,
			ChosenName:                    event.ChosenName,
			LastName:                    event.LastName,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.RenamedAt),
			ID:                   event.ID,
		})
	})
}

func (m *ReadModel) DeleteStudent(ctx context.Context, event StudentDeletedProjection) error {
	deletedAt := appdb.SQLTime(event.DeletedAt)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteStudent(conn, dbsql.DeleteStudentParams{
			DeletedAt:                &deletedAt,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			ID:                   event.ID,
		})
	})
}

func parseTime(value any) time.Time {
	text, _ := value.(string)
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return time.Now()
	}
	return parsed
}

func parseDBTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}
	return time.Time{}
}
