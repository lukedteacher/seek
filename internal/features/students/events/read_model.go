package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// student read model reader functions

func (m *ReadModel) GetByID(ctx context.Context, studentID string) (*models.Student, error) {
	var row *dbsql.GetStudentByIdRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudentById(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("student not found")
	}

	student := &models.Student{
		ID: row.Id,
		Person: sharedmodels.Person{
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:       sharedmodels.Grade(row.Grade),
		Homeroom:    row.Homeroom,
		CaseManager: row.CaseManager,
	}

	return student, nil
}
func (m *ReadModel) GetByUsername(ctx context.Context, username string) (*models.Student, error) {
	var row *dbsql.GetStudentByUsernameRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudentByUsername(conn, username)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("student not found")
	}

	student := &models.Student{
		ID: row.Id,
		Person: sharedmodels.Person{
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:       sharedmodels.Grade(row.Grade),
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
			ID: rows[i].Id,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
				Username:   rows[i].Username,
			},
			Grade:       sharedmodels.Grade(rows[i].Grade),
			Homeroom:    rows[i].Homeroom,
			CaseManager: rows[i].CaseManager,
		}
	}
	return students, nil
}

func (m *ReadModel) ListByIEPServiceType(ctx context.Context, serviceType string) ([]models.Student, error) {
	var rows []dbsql.ListStudentsByIepserviceTypeRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentsByIepserviceType(conn, serviceType)
		return err
	}); err != nil {
		return nil, err
	}

	students := make([]models.Student, len(rows))
	for i := range rows {
		students[i] = models.Student{
			ID: rows[i].Id,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
				Username:   rows[i].Username,
			},
			Grade:       sharedmodels.Grade(rows[i].Grade),
			Homeroom:    rows[i].Homeroom,
			CaseManager: rows[i].CaseManager,
		}
	}
	return students, nil
}

// student read model writer functions

func (m *ReadModel) Create(ctx context.Context, event StudentCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateStudent(conn, dbsql.CreateStudentParams{
			Id:                       event.StudentID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) Update(ctx context.Context, event StudentUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateStudent(conn, dbsql.UpdateStudentParams{
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			Id:                       event.StudentID,
		})
	})
}

func (m *ReadModel) Archive(ctx context.Context, event StudentArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveStudent(conn, dbsql.ArchiveStudentParams{
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			Id:                       event.StudentID,
		})
	})
}

func (m *ReadModel) Delete(ctx context.Context, event StudentDeletedProjection) error {
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
