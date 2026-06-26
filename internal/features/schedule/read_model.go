package schedule

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/domain/models"
	"seek/internal/uuidv7"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) Get(ctx context.Context, id string) (*models.Schedule, error) {
	var row *dbsql.GetScheduleRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetSchedule(conn, id)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	schedule := &models.Schedule{
		Id:        row.Id,
		Title:     row.Title,
		TeacherId: row.TeacherId,
	}

	return schedule, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Schedule, error) {
	var rows []dbsql.ListSchedulesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListSchedules(conn)
		return err
	}); err != nil {
		return nil, err
	}
	schedules := make([]models.Schedule, 0, len(rows))
	for _, row := range rows {
		schedules = append(schedules, models.Schedule{
			Id:        row.Id,
			Title:     row.Title,
			TeacherId: row.TeacherId,
		})
	}
	return schedules, nil
}

func (m *ReadModel) InsertCreatedSchedule(ctx context.Context, event ScheduleCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceInsertCreatedSchedule(conn, dbsql.InsertCreatedScheduleParams{
			Id:                       event.ScheduleID,
			Title:                    event.Title,
			TeacherId:                event.TeacherId,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateSchedule(ctx context.Context, event ScheduleUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateSchedule(conn, dbsql.UpdateScheduleParams{
			Id:                       event.ScheduleID,
			Title:                    event.Title,
			TeacherId:                event.TeacherId,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) PeriodAddedToSchedule(ctx context.Context, event PeriodAddedToScheduleProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceInsertCreatedSchedulePeriod(conn, dbsql.InsertCreatedSchedulePeriodParams{
			Id: 											uuidv7.NewString(),
			ScheduleId:               event.ScheduleID,
			PeriodId:                 event.PeriodID,
			CreatedAt:                appdb.SQLTime(event.PeriodAddedToScheduleAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) DeleteSchedule(ctx context.Context, event ScheduleDeletedProjection) error {
	deletedAt := appdb.SQLTime(event.DeletedAt)
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteSchedule(conn, dbsql.DeleteScheduleParams{
			DeletedAt:                &deletedAt,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			Id:                       event.ScheduleID,
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
