package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

func (m *ReadModel) Get(ctx context.Context, educatorID string) (*models.Educator, error) {
	var row *dbsql.GetEducatorRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetEducator(conn, educatorID)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("educator not found")
	}

	educator := &models.Educator{
		ID: row.Id,
		Person: sharedmodels.Person{
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Email:      row.Email,
		},
		Role: row.Role,
	}

	return educator, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Educator, error) {
	var rows []dbsql.ListEducatorsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListEducators(conn)
		return err
	}); err != nil {
		return nil, err
	}

	educators := make([]models.Educator, len(rows))
	for i := range rows {
		educators[i] = models.Educator{
			ID: rows[i].Id,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
			},
			Role: rows[i].Role,
		}
	}
	return educators, nil
}

func (m *ReadModel) CreateEducator(ctx context.Context, event EducatorCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateEducator(conn, dbsql.CreateEducatorParams{
			Id:                       event.ID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Role:                     event.Role,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateEducator(ctx context.Context, event EducatorUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateEducator(conn, dbsql.UpdateEducatorParams{
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Role:                     event.Role,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			Id:                       event.ID,
		})
	})
}

func (m *ReadModel) ArchiveEducator(ctx context.Context, event EducatorArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveEducator(conn, dbsql.ArchiveEducatorParams{
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			Id:                       event.ID,
		})
	})
}

func (m *ReadModel) DeleteEducator(ctx context.Context, event EducatorDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteEducator(conn, event.ID)
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
