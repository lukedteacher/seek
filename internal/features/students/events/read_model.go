package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/domain/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) Get(ctx context.Context, studentID string) (*models.Student, error) {
	var row *dbsql.GetStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudent(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("student not found")
	}

	student := &models.Student{
		ID:          row.Id,
		FirstName:   row.FirstName,
		ChosenName:  row.ChosenName,
		LastName:    row.LastName,
		Grade:       row.Grade,
		Homeroom:    row.Homeroom,
		CaseManager: row.CaseManager,
	}

	return student, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Student, error) {
	var rows []dbsql.ListStudentsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudents(conn)
		return err
	}); err != nil {
		return nil, err
	}
	
	students := make([]models.Student, len(rows))
	for i := range rows {
		students[i] = models.Student{
			ID:          rows[i].Id,
			FirstName:   rows[i].FirstName,
			ChosenName:  rows[i].ChosenName,
			LastName:    rows[i].LastName,
			Grade:       rows[i].Grade,
			Homeroom:    rows[i].Homeroom,
			CaseManager: rows[i].CaseManager,
		}
	}
	return students, nil
}

func (m *ReadModel) CreateStudent(ctx context.Context, event StudentCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateStudent(conn, dbsql.CreateStudentParams{
			Id:                       event.StudentID,
			FirstName:                event.FirstName,
			ChosenName:               event.ChosenName,
			LastName:                 event.LastName,
			Grade:                    event.Grade,
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateStudent(ctx context.Context, event StudentUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateStudent(conn, dbsql.UpdateStudentParams{
			FirstName:                event.FirstName,
			ChosenName:               event.ChosenName,
			LastName:                 event.LastName,
			Grade:                    event.Grade,
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			Id:                       event.StudentID,
		})
	})
}

func (m *ReadModel) DeleteStudent(ctx context.Context, event StudentDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteStudent(conn, event.StudentID)
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
