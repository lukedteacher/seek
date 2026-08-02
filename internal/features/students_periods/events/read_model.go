package events

import (
	"context"
	"fmt"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/students_periods/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

type PeriodStudentReadModelReader interface {
	Get(ctx context.Context, periodID, studentID string) (*models.PeriodStudent, error)
	List(ctx context.Context) ([]models.PeriodStudent, error)
	ListPeriodIDsForStudent(ctx context.Context, studentID string) ([]string, error)
	ListStudentIDsForPeriod(ctx context.Context, periodID string) ([]string, error)
}

type PeriodStudentReadModelWriter interface {
	AddStudentToPeriod(ctx context.Context, event StudentAddedToPeriodProjection) error
	RemoveStudentFromPeriod(ctx context.Context, event StudentRemovedFromPeriodProjection) error
}

// period student read model reader functions

func (m *ReadModel) Get(ctx context.Context, period_id, student_id string) (*models.PeriodStudent, error) {
	var row *dbsql.GetPeriodStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetPeriodStudent(conn, dbsql.GetPeriodStudentParams{
			PeriodId:  period_id,
			StudentId: student_id,
		})
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	return &models.PeriodStudent{
		PeriodID:  row.PeriodId,
		StudentID: row.StudentId,
	}, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.PeriodStudent, error) {
	var rows []dbsql.ListPeriodsStudentsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodsStudents(conn)
		return err
	}); err != nil {
		return nil, err
	}
	periodStudents := make([]models.PeriodStudent, len(rows))
	for i := range rows {
		periodStudents[i] = models.PeriodStudent{
			PeriodID:  rows[i].PeriodId,
			StudentID: rows[i].StudentId,
		}
	}
	return periodStudents, nil
}

func (m *ReadModel) ListPeriodIDsForStudent(ctx context.Context, student_id string) ([]string, error) {
	var rows []dbsql.ListPeriodIdsForStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodIdsForStudent(conn, student_id)
		return err
	}); err != nil {
		return nil, err
	}
	periodIDs := make([]string, len(rows))
	for i := range rows {
		periodIDs[i] = rows[i].PeriodId
	}
	return periodIDs, nil
}

func (m *ReadModel) ListStudentIDsForPeriod(ctx context.Context, period_id string) ([]string, error) {
	var rows []dbsql.ListStudentIdsForPeriodRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentIdsForPeriod(conn, period_id)
		return err
	}); err != nil {
		return nil, err
	}
	studentIDs := make([]string, len(rows))
	for i := range rows {
		studentIDs[i] = rows[i].StudentId
	}
	return studentIDs, nil
}

// period student read model writer functions

func (m *ReadModel) AddStudentToPeriod(ctx context.Context, event StudentAddedToPeriodProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddStudentToPeriod(conn, dbsql.AddStudentToPeriodParams{
			PeriodId:                 event.PeriodID,
			StudentId:                event.StudentID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemoveStudentFromPeriod(ctx context.Context, event StudentRemovedFromPeriodProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveStudentFromPeriod(conn, dbsql.RemoveStudentFromPeriodParams{
			PeriodId:  event.PeriodID,
			StudentId: event.StudentID,
		})
	})
}
