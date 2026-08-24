package events

import (
	"context"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
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

func (m *ReadModel) Get(ctx context.Context, studentID string) (*models.IEP, error) {
	var row *dbsql.GetIepRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetIep(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, appdb.ErrNoRows
	}

	iep := &models.IEP{}

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
	for i := range rows {
		ieps[i] = models.IEP{}
	}
	return ieps, nil
}

// read model writer functions

func (m *ReadModel) AddIEPToStudent(ctx context.Context, event IEPAddedToStudentProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddIeptoStudent(conn, dbsql.AddIeptoStudentParams{
			Id:          event.IEP.ID,
			StudentId:   event.IEP.StudentID,
			StartDate:   event.IEP.StartDate,
			EndDate:     event.IEP.EndDate,
			AmendedDate: event.IEP.AmendedDate,
			CreatedAt:   appdb.SQLTime(event.IEP.AddedAt),
		})
	})
}

func (m *ReadModel) UpdateIEP(ctx context.Context, event IEPUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateIep(conn, dbsql.UpdateIepParams{
			Id:          event.IEP.ID,
			StartDate:   event.IEP.StartDate,
			EndDate:     event.IEP.EndDate,
			AmendedDate: event.IEP.AmendedDate,
			UpdatedAt:   appdb.SQLTime(event.IEP.UpdatedAt),
		})
	})
}

func (m *ReadModel) DeleteIEP(ctx context.Context, event IEPDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteIep(conn, event.IEP.ID)
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
