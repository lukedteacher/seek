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
		ID:                    row.Id,
		StudentID:             row.StudentId,
		Disability1:           models.DisabilityCode(row.Disability1),
		Disability2:           models.DisabilityCode(row.Disability2),
		FederalSetting:        int(row.FederalSetting),
		MeetingDate:           sharedmodels.DateOnly(parseDBTime(row.MeetingDate)),
		IEPDueDate:            sharedmodels.DateOnly(parseDBTime(row.IepDueDate)),
		LastEvalDate:          sharedmodels.DateOnly(parseDBTime(row.LastEvalDate)),
		EvalDueDate:           sharedmodels.DateOnly(parseDBTime(row.EvalDueDate)),
		AmendedDate:           sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
		IEPType:               models.IEPType(row.IepType),
		SpecialTransportation: int64ToBool(row.SpecialTransportation),
		CreatedAt:             parseDBTime(row.CreatedAt),
		UpdatedAt:             parseDBTime(row.UpdatedAt),
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
			ID:                    row.Id,
			StudentID:             row.StudentId,
			Disability1:           models.DisabilityCode(row.Disability1),
			Disability2:           models.DisabilityCode(row.Disability2),
			FederalSetting:        int(row.FederalSetting),
			MeetingDate:           sharedmodels.DateOnly(parseDBTime(row.MeetingDate)),
			IEPDueDate:            sharedmodels.DateOnly(parseDBTime(row.IepDueDate)),
			LastEvalDate:          sharedmodels.DateOnly(parseDBTime(row.LastEvalDate)),
			EvalDueDate:           sharedmodels.DateOnly(parseDBTime(row.EvalDueDate)),
			AmendedDate:           sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
			IEPType:               models.IEPType(row.IepType),
			SpecialTransportation: int64ToBool(row.SpecialTransportation),
			CreatedAt:             parseDBTime(row.CreatedAt),
			UpdatedAt:             parseDBTime(row.UpdatedAt)}
	}
	return ieps, nil
}

func (m *ReadModel) ListForStudent(ctx context.Context, studentID string) ([]models.IEP, error) {
	var rows []dbsql.ListIepsForStudentRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListIepsForStudent(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	ieps := make([]models.IEP, len(rows))
	for i, row := range rows {
		ieps[i] = models.IEP{
			ID:                    row.Id,
			StudentID:             row.StudentId,
			Disability1:           models.DisabilityCode(row.Disability1),
			Disability2:           models.DisabilityCode(row.Disability2),
			FederalSetting:        int(row.FederalSetting),
			MeetingDate:           sharedmodels.DateOnly(parseDBTime(row.MeetingDate)),
			IEPDueDate:            sharedmodels.DateOnly(parseDBTime(row.IepDueDate)),
			LastEvalDate:          sharedmodels.DateOnly(parseDBTime(row.LastEvalDate)),
			EvalDueDate:           sharedmodels.DateOnly(parseDBTime(row.EvalDueDate)),
			AmendedDate:           sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
			IEPType:               models.IEPType(row.IepType),
			SpecialTransportation: int64ToBool(row.SpecialTransportation),
			CreatedAt:             parseDBTime(row.CreatedAt),
			UpdatedAt:             parseDBTime(row.UpdatedAt)}
	}
	return ieps, nil
}

// read model writer functions

func (m *ReadModel) AddIEPToStudent(ctx context.Context, event IEPAddedToStudentProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceAddIeptoStudent(conn, dbsql.AddIeptoStudentParams{
			Id:                       event.IEP.ID,
			StudentId:                event.IEP.StudentID,
			Disability1:              int64(event.Disability1),
			Disability2:              int64(event.Disability2),
			FederalSetting:           int64(event.FederalSetting),
			MeetingDate:              event.MeetingDate.String(),
			IepDueDate:               event.IEPDueDate.String(),
			LastEvalDate:             event.LastEvalDate.String(),
			EvalDueDate:              event.EvalDueDate.String(),
			AmendedDate:              event.IEP.AmendedDate.String(),
			IepType:                  int64(event.IEPType),
			SpecialTransportation:    boolToInt64(event.SpecialTransportation),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.IEP.CreatedAt),
		})
	})
}

func (m *ReadModel) UpdateIEP(ctx context.Context, event IEPUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateIep(conn, dbsql.UpdateIepParams{
			Id:                       event.IEP.ID,
			Disability1:              int64(event.Disability1),
			Disability2:              int64(event.Disability2),
			FederalSetting:           int64(event.FederalSetting),
			MeetingDate:              event.MeetingDate.String(),
			IepDueDate:               event.IEPDueDate.String(),
			LastEvalDate:             event.LastEvalDate.String(),
			EvalDueDate:              event.EvalDueDate.String(),
			AmendedDate:              event.IEP.AmendedDate.String(),
			IepType:                  int64(event.IEPType),
			SpecialTransportation:    boolToInt64(event.SpecialTransportation),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.IEP.UpdatedAt),
		})
	})
}

func (m *ReadModel) ArchiveIEP(ctx context.Context, event IEPArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveIep(conn, dbsql.ArchiveIepParams{
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) DeleteIEP(ctx context.Context, event IEPDeletedProjection) error {
	println("DELETE")
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteIep(conn, event.IEPID)
	})
}

func int64ToBool(i int64) bool {
	if i == 1 {
		return true
	}
	return false
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
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
