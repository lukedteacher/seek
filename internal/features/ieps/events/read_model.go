package events

import (
	"context"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/ieps/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// read model reader functions

func (m *ReadModel) Get(ctx context.Context, id string) (*models.IEP, error) {
	var row *dbsql.GetIepRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetIep(conn, id)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, appdb.ErrNoRows
	}

	iep := &models.IEP{
		ID:          row.Id,
		StudentID:   row.StudentId,
		StartDate:   sharedmodels.DateOnly(parseDBTime(row.StartDate)),
		EndDate:     sharedmodels.DateOnly(parseDBTime(row.EndDate)),
		AmendedDate: sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
		CreatedAt:   parseDBTime(row.CreatedAt),
		UpdatedAt:   parseDBTime(row.UpdatedAt),
	}

	return iep, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.IEP, error) {
	var rows []dbsql.ListIepsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListIeps(conn)
		return err
	}); err != nil {
		return nil, err
	}

	ieps := make([]models.IEP, len(rows))
	for i, row := range rows {
		ieps[i] = models.IEP{
			ID:          row.Id,
			StudentID:   row.StudentId,
			StartDate:   sharedmodels.DateOnly(parseDBTime(row.StartDate)),
			EndDate:     sharedmodels.DateOnly(parseDBTime(row.EndDate)),
			AmendedDate: sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
			CreatedAt:   parseDBTime(row.CreatedAt),
			UpdatedAt:   parseDBTime(row.UpdatedAt)}
	}
	return ieps, nil
}

// read model writer functions

func (m *ReadModel) AddIEPToStudent(ctx context.Context, event IEPAddedToStudentProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddIeptoStudent(conn, dbsql.AddIeptoStudentParams{
			Id:          event.IEPState.ID,
			StudentId:   event.IEPState.StudentID,
			StartDate:   event.IEPState.StartDate,
			EndDate:     event.IEPState.EndDate,
			AmendedDate: event.IEPState.AmendedDate,
			CreatedAt:   appdb.SQLTime(event.IEPState.AddedAt),
		})
	})
}

func (m *ReadModel) UpdateIEP(ctx context.Context, event IEPUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateIep(conn, dbsql.UpdateIepParams{
			Id:          event.IEPState.ID,
			StartDate:   event.IEPState.StartDate,
			EndDate:     event.IEPState.EndDate,
			AmendedDate: event.IEPState.AmendedDate,
			UpdatedAt:   appdb.SQLTime(event.IEPState.UpdatedAt),
		})
	})
}

func (m *ReadModel) DeleteIEP(ctx context.Context, event IEPDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteIep(conn, event.IEPState.ID)
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
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}
	return time.Time{}
}
