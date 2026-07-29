package events

import (
	"context"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/iepservices/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// read model reader functions

func (m *ReadModel) Get(ctx context.Context, id string) (*models.IEPService, error) {
	var row *dbsql.GetIepserviceRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetIepservice(conn, id)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, appdb.ErrNoRows
	}

	service := &models.IEPService{
		ID:              row.Id,
		StudentID:       row.StudentId,
		ServiceType:     sharedmodels.ServiceType(row.ServiceType),
		IndirectMinutes: int(row.IndirectMinutes),
		DirectMinutes:   int(row.DirectMinutes),
		FrequencyCount:  int(row.FrequencyCount),
		FrequencyType:   row.FrequencyType,
		Location:        row.Location,
		StartDate:       sharedmodels.DateOnly(parseDBTime(row.StartDate)),
		EndDate:         sharedmodels.DateOnly(parseDBTime(row.EndDate)),
		Provider:        row.Provider,
		CreatedAt:       parseDBTime(row.CreatedAt),
		UpdatedAt:       parseDBTime(row.UpdatedAt),
	}

	return service, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.IEPService, error) {
	var rows []dbsql.ListIepservicesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListIepservices(conn)
		return err
	}); err != nil {
		return nil, err
	}

	services := make([]models.IEPService, len(rows))
	for i := range rows {
		services[i] = models.IEPService{
			ID:              rows[i].Id,
			StudentID:       rows[i].StudentId,
			ServiceType:     sharedmodels.ServiceType(rows[i].ServiceType),
			IndirectMinutes: int(rows[i].IndirectMinutes),
			DirectMinutes:   int(rows[i].DirectMinutes),
			FrequencyCount:  int(rows[i].FrequencyCount),
			FrequencyType:   rows[i].FrequencyType,
			Location:        rows[i].Location,
			StartDate:       sharedmodels.DateOnly(parseDBTime(rows[i].StartDate)),
			EndDate:         sharedmodels.DateOnly(parseDBTime(rows[i].EndDate)),
			Provider:        rows[i].Provider,
			CreatedAt:       parseDBTime(rows[i].CreatedAt),
			UpdatedAt:       parseDBTime(rows[i].UpdatedAt),
		}
	}
	return services, nil
}

func (m *ReadModel) ListIEPServicesForStudent(ctx context.Context, studentID string) ([]models.IEPService, error) {
	var rows []dbsql.ListIepservicesForStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListIepservicesForStudent(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	services := make([]models.IEPService, len(rows))
	for i := range rows {
		services[i] = models.IEPService{
			ID:              rows[i].Id,
			StudentID:       rows[i].StudentId,
			ServiceType:     sharedmodels.ServiceType(rows[i].ServiceType),
			IndirectMinutes: int(rows[i].IndirectMinutes),
			DirectMinutes:   int(rows[i].DirectMinutes),
			FrequencyCount:  int(rows[i].FrequencyCount),
			FrequencyType:   rows[i].FrequencyType,
			Location:        rows[i].Location,
			StartDate:       sharedmodels.DateOnly(parseDBTime(rows[i].StartDate)),
			EndDate:         sharedmodels.DateOnly(parseDBTime(rows[i].EndDate)),
			Provider:        rows[i].Provider,
			CreatedAt:       parseDBTime(rows[i].CreatedAt),
			UpdatedAt:       parseDBTime(rows[i].UpdatedAt),
		}
	}
	return services, nil
}

// read model writer functions

func (m *ReadModel) AddIEPServiceToStudent(ctx context.Context, event IEPServiceAddedToStudentProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddIepserviceToStudent(conn, dbsql.AddIepserviceToStudentParams{
			Id:              event.IEPServiceID,
			StudentId:       event.StudentID,
			ServiceType:     event.ServiceType,
			IndirectMinutes: int64(event.IndirectMinutes),
			DirectMinutes:   int64(event.DirectMinutes),
			FrequencyCount:  int64(event.FrequencyCount),
			FrequencyType:   event.FrequencyType,
			Location:        event.Location,
			StartDate:       event.StartDate,
			EndDate:         event.EndDate,
			Provider:        event.Provider,
			CreatedAt:       appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateIEPService(ctx context.Context, event IEPServiceUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateIepservice(conn, dbsql.UpdateIepserviceParams{
			Id:              event.IEPServiceID,
			StudentId:       event.StudentID,
			ServiceType:     event.ServiceType,
			IndirectMinutes: int64(event.IndirectMinutes),
			DirectMinutes:   int64(event.DirectMinutes),
			FrequencyCount:  int64(event.FrequencyCount),
			FrequencyType:   event.FrequencyType,
			Location:        event.Location,
			StartDate:       event.StartDate,
			EndDate:         event.EndDate,
			Provider:        event.Provider,
			UpdatedAt:       appdb.SQLTime(event.UpdatedAt),
		})
	})
}

func (m *ReadModel) DeleteIEPService(ctx context.Context, event IEPServiceDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteIepservice(conn, event.IEPServiceID)
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
