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

type PeriodScheduleReadModelReader interface {
	Get(ctx context.Context, periodID, studentID string) (*models.PeriodSchedule, error)
	List(ctx context.Context) ([]models.PeriodSchedule, error)
	ListPeriodIDsForSchedule(ctx context.Context, studentID string) ([]string, error)
	ListScheduleIDsForPeriod(ctx context.Context, periodID string) ([]string, error)
}

type PeriodScheduleReadModelWriter interface {
	AddPeriodToSchedule(ctx context.Context, event PeriodScheduleAddedProjection) error
	RemovePeriodFromSchedule(ctx context.Context, event PeriodScheduleRemovedProjection) error
}

// READ MODEL READER FUNCTIONS

func (m *ReadModel) Get(ctx context.Context, period_id, student_id string) (*models.PeriodSchedule, error) {
	var row *dbsql.GetPeriodScheduleRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetPeriodSchedule(conn, dbsql.GetPeriodScheduleParams{
			PeriodId:   period_id,
			ScheduleId: student_id,
		})
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	return &models.PeriodSchedule{
		PeriodID:   row.PeriodId,
		ScheduleID: row.ScheduleId,
	}, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.PeriodSchedule, error) {
	var rows []dbsql.ListPeriodsSchedulesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodsSchedules(conn)
		return err
	}); err != nil {
		return nil, err
	}
	periodSchedules := make([]models.PeriodSchedule, len(rows))
	for i := range rows {
		periodSchedules[i] = models.PeriodSchedule{
			PeriodID:   rows[i].PeriodId,
			ScheduleID: rows[i].ScheduleId,
		}
	}
	return periodSchedules, nil
}

func (m *ReadModel) ListPeriodIDsForSchedule(ctx context.Context, student_id string) ([]string, error) {
	var rows []dbsql.ListPeriodIdsForScheduleRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodIdsForSchedule(conn, student_id)
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

func (m *ReadModel) ListScheduleIDsForPeriod(ctx context.Context, period_id string) ([]string, error) {
	var rows []dbsql.ListScheduleIdsForPeriodRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListScheduleIdsForPeriod(conn, period_id)
		return err
	}); err != nil {
		return nil, err
	}
	studentIDs := make([]string, len(rows))
	for i := range rows {
		studentIDs[i] = rows[i].ScheduleId
	}
	return studentIDs, nil
}

// READ MODEL WRITER FUNCTIONS

func (m *ReadModel) AddPeriodToSchedule(ctx context.Context, event PeriodScheduleAddedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddPeriodToSchedule(conn, dbsql.AddPeriodToScheduleParams{
			PeriodId:                 event.PeriodID,
			ScheduleId:               event.ScheduleID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemovePeriodFromSchedule(ctx context.Context, event PeriodScheduleRemovedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemovePeriodFromSchedule(conn, dbsql.RemovePeriodFromScheduleParams{
			PeriodId:   event.PeriodID,
			ScheduleId: event.ScheduleID,
		})
	})
}

// HELPERS

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
