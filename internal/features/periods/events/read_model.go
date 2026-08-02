package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/periods/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// period read model reader functions

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
		ID:          row.Id,
		Title:       row.Title,
		ServiceType: sharedmodels.ServiceType(row.ServiceType),
		StartTime:   parseDBTimeOnly(row.StartTime),
		EndTime:     parseDBTimeOnly(row.StartTime).Add(int(row.Duration)),
		Duration:    int(row.Duration),
		DaysBitmask: sharedmodels.DaysBitmask(row.DaysBitmask),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
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
			ID:          row.Id,
			Title:       row.Title,
			ServiceType: sharedmodels.ServiceType(row.ServiceType),
			StartTime:   parseDBTimeOnly(row.StartTime),
			EndTime:     parseDBTimeOnly(row.StartTime).Add(int(row.Duration)),
			Duration:    int(row.Duration),
			DaysBitmask: sharedmodels.DaysBitmask(row.DaysBitmask),
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return periods, nil
}

func (m *ReadModel) ListPeriodsForEducator(ctx context.Context, educatorID string) ([]models.Period, error) {
	var rows []dbsql.ListPeriodsForEducatorRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodsForEducator(conn, educatorID)
		return err
	}); err != nil {
		return nil, err
	}
	periods := make([]models.Period, 0, len(rows))
	for _, row := range rows {
		periods = append(periods, models.Period{
			ID:          row.Id,
			Title:       row.Title,
			ServiceType: sharedmodels.ServiceType(row.ServiceType),
			StartTime:   parseDBTimeOnly(row.StartTime),
			EndTime:     parseDBTimeOnly(row.StartTime).Add(int(row.Duration)),
			Duration:    int(row.Duration),
			DaysBitmask: sharedmodels.DaysBitmask(row.DaysBitmask),
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return periods, nil
}

func (m *ReadModel) ListPeriodsForStudent(ctx context.Context, studentID string) ([]models.Period, error) {
	var rows []dbsql.ListPeriodsForStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodsForStudent(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}
	periods := make([]models.Period, 0, len(rows))
	for _, row := range rows {
		periods = append(periods, models.Period{
			ID:          row.Id,
			Title:       row.Title,
			ServiceType: sharedmodels.ServiceType(row.ServiceType),
			StartTime:   parseDBTimeOnly(row.StartTime),
			EndTime:     parseDBTimeOnly(row.StartTime).Add(int(row.Duration)),
			Duration:    int(row.Duration),
			DaysBitmask: sharedmodels.DaysBitmask(row.DaysBitmask),
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return periods, nil
}

func (m *ReadModel) CreatePeriod(ctx context.Context, event PeriodCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreatePeriod(conn, dbsql.CreatePeriodParams{
			Id:                       event.PeriodID,
			Title:                    event.Title,
			ServiceType:              string(event.ServiceType),
			StartTime:                event.StartTime.String(),
			Duration:                 int64(event.Duration),
			DaysBitmask:              int64(event.DaysBitmask),
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdatePeriod(ctx context.Context, event PeriodUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdatePeriod(conn, dbsql.UpdatePeriodParams{
			Id:                       event.PeriodID,
			Title:                    event.Title,
			ServiceType:              string(event.ServiceType),
			StartTime:                event.StartTime.String(),
			Duration:                 int64(event.Duration),
			DaysBitmask:              int64(event.DaysBitmask),
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) ArchivePeriod(ctx context.Context, event PeriodArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchivePeriod(conn, dbsql.ArchivePeriodParams{
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			Id:                       event.PeriodID,
		})
	})
}

func (m *ReadModel) DeletePeriod(ctx context.Context, event PeriodDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeletePeriod(conn, event.PeriodID)
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

func parseDBTimeOnly(value string) sharedmodels.TimeOnly {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return sharedmodels.TimeOnly(parsed)
		}
	}
	return sharedmodels.TimeOnly{}
}
