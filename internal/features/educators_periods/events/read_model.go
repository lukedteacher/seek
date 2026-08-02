package events

import (
	"context"
	"fmt"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/educators_periods/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

type PeriodEducatorReadModelReader interface {
	Get(ctx context.Context, periodID, educatorID string) (*models.PeriodEducator, error)
	List(ctx context.Context) ([]models.PeriodEducator, error)
	ListPeriodIDsForEducator(ctx context.Context, educatorID string) ([]string, error)
	ListEducatorIDsForPeriod(ctx context.Context, periodID string) ([]string, error)
}

type PeriodEducatorReadModelWriter interface {
	AddEducatorToPeriod(ctx context.Context, event EducatorAddedToPeriodProjection) error
	RemoveEducatorFromPeriod(ctx context.Context, event EducatorRemovedFromPeriodProjection) error
}

// period educator read model reader functions

func (m *ReadModel) Get(ctx context.Context, period_id, educator_id string) (*models.PeriodEducator, error) {
	var row *dbsql.GetPeriodEducatorRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetPeriodEducator(conn, dbsql.GetPeriodEducatorParams{
			PeriodId:   period_id,
			EducatorId: educator_id,
		})
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	return &models.PeriodEducator{
		PeriodID:   row.PeriodId,
		EducatorID: row.EducatorId,
	}, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.PeriodEducator, error) {
	var rows []dbsql.ListPeriodsEducatorsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodsEducators(conn)
		return err
	}); err != nil {
		return nil, err
	}
	periodEducators := make([]models.PeriodEducator, len(rows))
	for i := range rows {
		periodEducators[i] = models.PeriodEducator{
			PeriodID:   rows[i].PeriodId,
			EducatorID: rows[i].EducatorId,
		}
	}
	return periodEducators, nil
}

func (m *ReadModel) ListPeriodIDsForEducator(ctx context.Context, educator_id string) ([]string, error) {
	var rows []dbsql.ListPeriodIdsForEducatorRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListPeriodIdsForEducator(conn, educator_id)
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

func (m *ReadModel) ListEducatorIDsForPeriod(ctx context.Context, period_id string) ([]string, error) {
	var rows []dbsql.ListEducatorIdsForPeriodRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListEducatorIdsForPeriod(conn, period_id)
		return err
	}); err != nil {
		return nil, err
	}
	educatorIDs := make([]string, len(rows))
	for i := range rows {
		educatorIDs[i] = rows[i].EducatorId
	}
	return educatorIDs, nil
}

// period educator read model writer functions

func (m *ReadModel) AddEducatorToPeriod(ctx context.Context, event EducatorAddedToPeriodProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddEducatorToPeriod(conn, dbsql.AddEducatorToPeriodParams{
			PeriodId:                 event.PeriodID,
			EducatorId:               event.EducatorID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemoveEducatorFromPeriod(ctx context.Context, event EducatorRemovedFromPeriodProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveEducatorFromPeriod(conn, dbsql.RemoveEducatorFromPeriodParams{
			PeriodId:   event.PeriodID,
			EducatorId: event.EducatorID,
		})
	})
}
