package events

import (
	"context"
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

// READ MODEL READER FUNCTIONS

func (m *ReadModel) Get(ctx context.Context, id string) (*models.StudentService, error) {
	var row *dbsql.GetStudentServiceRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudentService(conn, id)
		return err
	}); err != nil {
		return nil, err
	}
	if row == nil {
		return nil, appdb.ErrNoRows
	}

	// convert nullable pointer fields with safe defaults
	indirectMinutes := 0
	if row.IndirectMinutes != nil {
		indirectMinutes = int(*row.IndirectMinutes)
	}
	directMinutes := 0
	if row.DirectMinutes != nil {
		directMinutes = int(*row.DirectMinutes)
	}
	frequencyCount := 0
	if row.FrequencyCount != nil {
		frequencyCount = int(*row.FrequencyCount)
	}

	frequencyType := ""
	if row.FrequencyType != nil {
		frequencyType = *row.FrequencyType
	}
	location := ""
	if row.Location != nil {
		location = *row.Location
	}
	startDate := ""
	if row.StartDate != nil {
		startDate = *row.StartDate
	}
	endDate := ""
	if row.EndDate != nil {
		endDate = *row.EndDate
	}
	provider := ""
	if row.Provider != nil {
		provider = *row.Provider
	}

	service := &models.StudentService{
		ServiceID:       row.Id,
		StudentID:       row.StudentId,
		ServiceType:     row.ServiceType,
		IndirectMinutes: indirectMinutes,
		DirectMinutes:   directMinutes,
		FrequencyCount:  frequencyCount,
		FrequencyType:   frequencyType,
		Location:        location,
		StartDate:       startDate,
		EndDate:         endDate,
		Provider:        provider,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}

	return service, nil
}

func (m *ReadModel) List(ctx context.Context) ([]models.StudentService, error) {
	var rows []dbsql.ListStudentServicesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentServices(conn)
		return err
	}); err != nil {
		return nil, err
	}

	services := make([]models.StudentService, len(rows))
	for i := range rows {
		// convert nullable pointer fields with safe defaults
		indirectMinutes := 0
		if rows[i].IndirectMinutes != nil {
			indirectMinutes = int(*rows[i].IndirectMinutes)
		}
		directMinutes := 0
		if rows[i].DirectMinutes != nil {
			directMinutes = int(*rows[i].DirectMinutes)
		}
		frequencyCount := 0
		if rows[i].FrequencyCount != nil {
			frequencyCount = int(*rows[i].FrequencyCount)
		}

		frequencyType := ""
		if rows[i].FrequencyType != nil {
			frequencyType = *rows[i].FrequencyType
		}
		location := ""
		if rows[i].Location != nil {
			location = *rows[i].Location
		}
		startDate := ""
		if rows[i].StartDate != nil {
			startDate = *rows[i].StartDate
		}
		endDate := ""
		if rows[i].EndDate != nil {
			endDate = *rows[i].EndDate
		}
		provider := ""
		if rows[i].Provider != nil {
			provider = *rows[i].Provider
		}

		services[i] = models.StudentService{
			ServiceID:       rows[i].Id,
			StudentID:       rows[i].StudentId,
			ServiceType:     rows[i].ServiceType,
			IndirectMinutes: indirectMinutes,
			DirectMinutes:   directMinutes,
			FrequencyCount:  frequencyCount,
			FrequencyType:   frequencyType,
			Location:        location,
			StartDate:       startDate,
			EndDate:         endDate,
			Provider:        provider,
			CreatedAt:       rows[i].CreatedAt,
			UpdatedAt:       rows[i].UpdatedAt,
		}
	}
	return services, nil
}

// READ MODEL WRITER FUNCTIONS

func (m *ReadModel) CreateStudentService(ctx context.Context, event StudentServiceCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		// convert int fields to int64 (SQL expects int64)
		indirectMinutes := int64(event.IndirectMinutes)
		directMinutes := int64(event.DirectMinutes)
		frequencyCount := int64(event.FrequencyCount)

		var location *string
		if event.Location != "" {
			location = &event.Location
		}
		var startDate *string
		if event.StartDate != "" {
			startDate = &event.StartDate
		}
		var endDate *string
		if event.EndDate != "" {
			endDate = &event.EndDate
		}
		var provider *string
		if event.Provider != "" {
			provider = &event.Provider
		}

		var frequencyType *string
		if event.FrequencyType != "" {
			frequencyType = &event.FrequencyType
		}

		return dbsql.OnceCreateStudentService(conn, dbsql.CreateStudentServiceParams{
			Id:              event.ServiceID,
			StudentId:       event.StudentID,
			ServiceType:     event.ServiceType,
			IndirectMinutes: &indirectMinutes,
			DirectMinutes:   &directMinutes,
			FrequencyCount:  &frequencyCount,
			FrequencyType:   frequencyType,
			Location:        location,
			StartDate:       startDate,
			EndDate:         endDate,
			Provider:        provider,
			CreatedAt:       appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateStudentService(ctx context.Context, event StudentServiceUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		// convert int fields to int64 (SQL expects *int64 for nullable columns)
		indirectMinutes := int64(event.IndirectMinutes)
		directMinutes := int64(event.DirectMinutes)
		frequencyCount := int64(event.FrequencyCount)

		// for nullable string fields, pass nil if empty
		var frequencyType *string
		if event.FrequencyType != "" {
			frequencyType = &event.FrequencyType
		}
		var location *string
		if event.Location != "" {
			location = &event.Location
		}
		var startDate *string
		if event.StartDate != "" {
			startDate = &event.StartDate
		}
		var endDate *string
		if event.EndDate != "" {
			endDate = &event.EndDate
		}
		var provider *string
		if event.Provider != "" {
			provider = &event.Provider
		}

		return dbsql.OnceUpdateStudentService(conn, dbsql.UpdateStudentServiceParams{
			Id:              event.ServiceID,
			StudentId:       event.StudentID,
			ServiceType:     event.ServiceType,
			IndirectMinutes: &indirectMinutes,
			DirectMinutes:   &directMinutes,
			FrequencyCount:  &frequencyCount,
			FrequencyType:   frequencyType,
			Location:        location,
			StartDate:       startDate,
			EndDate:         endDate,
			Provider:        provider,
			UpdatedAt:       appdb.SQLTime(event.UpdatedAt),
		})
	})
}

func (m *ReadModel) DeleteStudentService(ctx context.Context, event StudentServiceDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteStudentService(conn, event.ServiceID)
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
