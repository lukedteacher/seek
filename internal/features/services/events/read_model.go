package events

import (
	"context"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/services/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// read model reader functions

func (m *ReadModel) Get(ctx context.Context, id string) (*models.Service, error) {
	var row *dbsql.GetServiceRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetService(conn, id)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, appdb.ErrNoRows
	}

	service := &models.Service{
		ID:              row.Id,
		IEPID:           row.IepId,
		ServiceName:     row.ServiceName,
		ServiceType:     sharedmodels.ServiceType(row.ServiceType),
		IndirectMinutes: int(row.IndirectMinutes),
		DirectMinutes:   int(row.DirectMinutes),
		FrequencyCount:  int(row.FrequencyCount),
		FrequencyType:   row.FrequencyType,
		LocationID:      row.LocationId,
		StartDate:       sharedmodels.DateOnly(parseDBTime(row.StartDate)),
		EndDate:         sharedmodels.DateOnly(parseDBTime(row.EndDate)),
		ProviderID:      row.ProviderId,
		CreatedAt:       parseDBTime(row.CreatedAt),
		UpdatedAt:       parseDBTime(row.UpdatedAt),
	}

	return service, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.Service, error) {
	var rows []dbsql.ListServicesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListServices(conn)
		return err
	}); err != nil {
		return nil, err
	}

	services := make([]models.Service, len(rows))
	for i := range rows {
		services[i] = models.Service{
			ID:              rows[i].Id,
			IEPID:           rows[i].IepId,
			ServiceName:     rows[i].ServiceName,
			ServiceType:     sharedmodels.ServiceType(rows[i].ServiceType),
			IndirectMinutes: int(rows[i].IndirectMinutes),
			DirectMinutes:   int(rows[i].DirectMinutes),
			FrequencyCount:  int(rows[i].FrequencyCount),
			FrequencyType:   rows[i].FrequencyType,
			LocationID:      rows[i].LocationId,
			StartDate:       sharedmodels.DateOnly(parseDBTime(rows[i].StartDate)),
			EndDate:         sharedmodels.DateOnly(parseDBTime(rows[i].EndDate)),
			Provider:        rows[i].ProviderId,
			CreatedAt:       parseDBTime(rows[i].CreatedAt),
			UpdatedAt:       parseDBTime(rows[i].UpdatedAt),
		}
	}
	return services, nil
}

func (m *ReadModel) ListServicesForIEP(ctx context.Context, studentID string) ([]models.Service, error) {
	var rows []dbsql.ListServicesForIepRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListServicesForIep(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	services := make([]models.Service, len(rows))
	for i := range rows {
		services[i] = models.Service{
			ID:              rows[i].Id,
			IEPID:           rows[i].IepId,
			ServiceName:     rows[i].ServiceName,
			ServiceType:     sharedmodels.ServiceType(rows[i].ServiceType),
			IndirectMinutes: int(rows[i].IndirectMinutes),
			DirectMinutes:   int(rows[i].DirectMinutes),
			FrequencyCount:  int(rows[i].FrequencyCount),
			FrequencyType:   rows[i].FrequencyType,
			LocationID:      rows[i].LocationId,
			StartDate:       sharedmodels.DateOnly(parseDBTime(rows[i].StartDate)),
			EndDate:         sharedmodels.DateOnly(parseDBTime(rows[i].EndDate)),
			Provider:        rows[i].ProviderId,
			CreatedAt:       parseDBTime(rows[i].CreatedAt),
			UpdatedAt:       parseDBTime(rows[i].UpdatedAt),
		}
	}
	return services, nil
}

// read model writer functions

func (m *ReadModel) AddServiceToIEP(ctx context.Context, event ServiceAddedToIEPProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddServiceToIep(conn, dbsql.AddServiceToIepParams{
			Id:              event.ServiceID,
			IepId:           event.IEPID,
			ServiceName:     event.ServiceName,
			ServiceType:     event.ServiceType,
			IndirectMinutes: int64(event.IndirectMinutes),
			DirectMinutes:   int64(event.DirectMinutes),
			FrequencyCount:  int64(event.FrequencyCount),
			FrequencyType:   event.FrequencyType,
			LocationId:      event.LocationID,
			StartDate:       event.StartDate,
			EndDate:         event.EndDate,
			ProviderId:      event.ProviderID,
			CreatedAt:       appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateService(ctx context.Context, event ServiceUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateService(conn, dbsql.UpdateServiceParams{
			Id:              event.ServiceID,
			ServiceName:     event.ServiceName,
			ServiceType:     event.ServiceType,
			IndirectMinutes: int64(event.IndirectMinutes),
			DirectMinutes:   int64(event.DirectMinutes),
			FrequencyCount:  int64(event.FrequencyCount),
			FrequencyType:   event.FrequencyType,
			LocationId:      event.LocationID,
			StartDate:       event.StartDate,
			EndDate:         event.EndDate,
			ProviderId:      event.ProviderID,
			UpdatedAt:       appdb.SQLTime(event.UpdatedAt),
		})
	})
}

func (m *ReadModel) DeleteService(ctx context.Context, event ServiceDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteService(conn, event.ServiceID)
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
