package period

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

func (m *ReadModel) Get(ctx context.Context, id string) (*models.Period, error) {
	var row *dbsql.GetPeriodRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetPeriod(conn, id)
		return err
	}); err != nil {
		return nil, err
	}
	
	if row == nil {
		return nil, fmt.Errorf("period not found")
	}
	
	period := &models.Period{
		Id:        row.Id,
		Title:     row.Title,
		StartTime: row.StartTime,
		Duration:  row.Duration,
		Days:      row.Days,
	}
	
	return period, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Period, error) {
	var rows []dbsql.ListPeriodsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriods(conn)
		return err
	}); err != nil {
		return nil, err
	}
	periods := make([]models.Period, 0, len(rows))
	for _, row := range rows {
		periods = append(periods, models.Period{
			Id:				 row.Id,
			Title:		 row.Title,
			StartTime: row.StartTime,
			Duration:  row.Duration,
			Days:      row.Days,
		})
	}
	return periods, nil
}

func (m *ReadModel) InsertCreatedPeriod(ctx context.Context, event PeriodCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceInsertCreatedPeriod(conn, dbsql.InsertCreatedPeriodParams{
			Id:                       event.Id,
			Title:                    event.Title,
			StartTime:                event.StartTime,
			Duration:                 event.Duration,
			Days:                     event.Days,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RenamePeriod(ctx context.Context, event PeriodRenamedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRenamePeriod(conn, dbsql.RenamePeriodParams{
			Title:                    event.Title,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.RenamedAt),
			Id:                   event.Id,
		})
	})
}

func (m *ReadModel) DeletePeriod(ctx context.Context, event PeriodDeletedProjection) error {
	deletedAt := appdb.SQLTime(event.DeletedAt)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeletePeriod(conn, dbsql.DeletePeriodParams{
			DeletedAt:                &deletedAt,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			Id:                   event.Id,
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
