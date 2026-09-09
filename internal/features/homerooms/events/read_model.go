package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/homerooms/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// homeroom read model reader functions

func (m *ReadModel) Get(ctx context.Context, homeroomID string) (*models.Homeroom, error) {
	var row *dbsql.GetHomeroomRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetHomeroom(conn, homeroomID)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("homeroom not found")
	}
	homeroom := &models.Homeroom{
		ID:            row.Id,
		Title:         row.Title,
		GradesBitmask: sharedmodels.GradesBitmask(row.GradesBitmask),
		LocationID:    row.LocationId,
		Image:         row.Image,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	return homeroom, nil
}

func (m *ReadModel) GetWithIDs(ctx context.Context, homeroomID string) (*models.Homeroom, error) {
	var row *dbsql.GetHomeroomWithIdsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetHomeroomWithIds(conn, homeroomID)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("homeroom not found")
	}
	var educatorIDs []string
	if err := json.Unmarshal([]byte(row.EducatorIds), &educatorIDs); err != nil {
		return nil, err
	}
	var studentIDs []string
	if err := json.Unmarshal([]byte(row.StudentIds), &studentIDs); err != nil {
		return nil, err
	}
	homeroom := &models.Homeroom{
		ID:            row.Id,
		Title:         row.Title,
		GradesBitmask: sharedmodels.GradesBitmask(row.GradesBitmask),
		LocationID:    row.LocationId,
		Image:         row.Image,
		EducatorIDs:   educatorIDs,
		StudentIDs:    studentIDs,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	return homeroom, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Homeroom, error) {
	var rows []dbsql.ListHomeroomsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListHomerooms(conn)
		return err
	}); err != nil {
		return nil, err
	}
	homerooms := make([]models.Homeroom, 0, len(rows))
	for _, row := range rows {
		homerooms = append(homerooms, models.Homeroom{
			ID:            row.Id,
			Title:         row.Title,
			GradesBitmask: sharedmodels.GradesBitmask(row.GradesBitmask),
			LocationID:    row.LocationId,
			Image:         row.Image,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return homerooms, nil
}

func (m *ReadModel) ListWithIDs(ctx context.Context) ([]models.Homeroom, error) {
	var rows []dbsql.ListHomeroomsWithIdsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListHomeroomsWithIds(conn)
		return err
	}); err != nil {
		return nil, err
	}
	homerooms := make([]models.Homeroom, 0, len(rows))
	for _, row := range rows {

		var educatorIDs []string
		if err := json.Unmarshal([]byte(row.EducatorIds), &educatorIDs); err != nil {
			return nil, err
		}
		var studentIDs []string
		if err := json.Unmarshal([]byte(row.StudentIds), &studentIDs); err != nil {
			return nil, err
		}
		homerooms = append(homerooms, models.Homeroom{
			ID:            row.Id,
			Title:         row.Title,
			GradesBitmask: sharedmodels.GradesBitmask(row.GradesBitmask),
			LocationID:    row.LocationId,
			Image:         row.Image,
			EducatorIDs:   educatorIDs,
			StudentIDs:    studentIDs,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}
	return homerooms, nil
}

// homeroom read model writer functions

func (m *ReadModel) CreateHomeroom(ctx context.Context, event HomeroomCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateHomeroom(conn, dbsql.CreateHomeroomParams{
			Id:                       event.Homeroom.ID,
			Title:                    event.Homeroom.Title,
			GradesBitmask:            int64(event.Homeroom.GradesBitmask),
			LocationId:               event.Homeroom.LocationID,
			Image:                    event.Homeroom.Image,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) UpdateHomeroom(ctx context.Context, event HomeroomUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateHomeroom(conn, dbsql.UpdateHomeroomParams{
			Id:                       event.Homeroom.ID,
			Title:                    event.Homeroom.Title,
			GradesBitmask:            int64(event.Homeroom.GradesBitmask),
			LocationId:               event.Homeroom.LocationID,
			Image:                    event.Homeroom.Image,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) ArchiveHomeroom(ctx context.Context, event HomeroomArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveHomeroom(conn, dbsql.ArchiveHomeroomParams{
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			Id:                       event.HomeroomID,
		})
	})
}

func (m *ReadModel) DeleteHomeroom(ctx context.Context, event HomeroomDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteHomeroom(conn, event.HomeroomID)
	})
}

func (m *ReadModel) AddEducatorToHomeroom(ctx context.Context, event EducatorAddedToHomeroomProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddEducatorToHomeroom(conn, dbsql.AddEducatorToHomeroomParams{
			HomeroomId:               event.HomeroomID,
			EducatorId:               event.EducatorID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemoveEducatorFromHomeroom(ctx context.Context, event EducatorRemovedFromHomeroomProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveEducatorFromHomeroom(conn, dbsql.RemoveEducatorFromHomeroomParams{
			HomeroomId: event.HomeroomID,
			EducatorId: event.EducatorID,
		})
	})
}

func (m *ReadModel) AddStudentToHomeroom(ctx context.Context, event StudentAddedToHomeroomProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddStudentToHomeroom(conn, dbsql.AddStudentToHomeroomParams{
			HomeroomId:               event.HomeroomID,
			StudentId:                event.StudentID,
			CreatedAt:                appdb.SQLTime(event.AddedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) RemoveStudentFromHomeroom(ctx context.Context, event StudentRemovedFromHomeroomProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceRemoveStudentFromHomeroom(conn, dbsql.RemoveStudentFromHomeroomParams{
			HomeroomId: event.HomeroomID,
			StudentId:  event.StudentID,
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

func parseDBTimeOnly(value string) sharedmodels.TimeOnly {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return sharedmodels.TimeOnly(parsed)
		}
	}
	return sharedmodels.TimeOnly{}
}
