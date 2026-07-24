package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/teachers/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) Get(ctx context.Context, teacherID string) (*models.Teacher, error) {
	var row *dbsql.GetTeacherRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetTeacher(conn, teacherID)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("teacher not found")
	}

	teacher := &models.Teacher{
		ID:         row.Id,
		FirstName:  row.FirstName,
		ChosenName: row.ChosenName,
		LastName:   row.LastName,
	}

	return teacher, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Teacher, error) {
	var rows []dbsql.ListTeachersRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListTeachers(conn)
		return err
	}); err != nil {
		return nil, err
	}
	teachers := make([]models.Teacher, 0, len(rows))
	for _, row := range rows {
		teachers = append(teachers, models.Teacher{
			ID:         row.Id,
			FirstName:  row.FirstName,
			ChosenName: row.ChosenName,
			LastName:   row.LastName,
		})
	}
	return teachers, nil
}

func (m *ReadModel) CreateTeacher(ctx context.Context, event TeacherCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateTeacher(conn, dbsql.CreateTeacherParams{
			Id:                       event.TeacherID,
			FirstName:                event.FirstName,
			ChosenName:               event.ChosenName,
			LastName:                 event.LastName,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateTeacher(ctx context.Context, event TeacherUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateTeacher(conn, dbsql.UpdateTeacherParams{
			FirstName:                event.FirstName,
			ChosenName:               event.ChosenName,
			LastName:                 event.LastName,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			Id:                       event.TeacherID,
		})
	})
}

func (m *ReadModel) DeleteTeacher(ctx context.Context, event TeacherDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteTeacher(conn, event.TeacherID)
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
