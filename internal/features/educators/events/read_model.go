package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	caseloadStudentsModels "seek/internal/features/caseload_students/models"
	"seek/internal/features/educators/models"
	studentModels "seek/internal/features/students/models"

	"zombiezen.com/go/sqlite"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// educator read model reader functions

type GetOption func(*getConfig)

type getConfig struct {
	withRoles    bool
	withCaseload bool
}

func WithCaseload() GetOption {
	return func(c *getConfig) {
		c.withCaseload = true
	}
}

func WithRoles() GetOption {
	return func(c *getConfig) {
		c.withRoles = true
	}
}

func (m *ReadModel) GetByUsername(
	ctx context.Context,
	username string,
	opts ...GetOption,
) (
	*models.Educator,
	error,
) {
	cfg := &getConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.withRoles {
		return m.getByUsernameWithRoles(ctx, username)
	}
	return m.getByUsername(ctx, username)
}

func (m *ReadModel) GetByID(ctx context.Context, educatorID string) (*models.Educator, error) {
	var row *dbsql.GetEducatorByIdRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetEducatorById(conn, educatorID)
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
			Username:   row.Username,
		},
	}

	return educator, nil
}

func (m *ReadModel) getByUsername(ctx context.Context, username string) (*models.Educator, error) {
	var row *dbsql.GetEducatorByUsernameRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetEducatorByUsername(conn, username)
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
			Username:   row.Username,
		},
	}

	return educator, nil
}

func (m *ReadModel) getByUsernameWithRoles(ctx context.Context, username string) (*models.Educator, error) {
	var rows []dbsql.GetEducatorWithRolesByUsernameRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceGetEducatorWithRolesByUsername(conn, username)
		return err
	}); err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("educator not found")
	}

	educator := &models.Educator{
		ID: rows[0].Id,
		Person: sharedmodels.Person{
			GivenName:  rows[0].GivenName,
			ChosenName: rows[0].ChosenName,
			FamilyName: rows[0].FamilyName,
			Email:      rows[0].Email,
			Username:   rows[0].Username,
		},
		Roles: make([]sharedmodels.EducatorRole, len(rows)),
	}
	for i, row := range rows {
		if row.Role != nil {
			educator.Roles[i] = sharedmodels.EducatorRole(*row.Role)
		}
	}

	return educator, nil
}

func (m *ReadModel) GetByUsernameWithCaseload(ctx context.Context, username string) (*caseloadStudentsModels.CaseManager, error) {
	var rows []dbsql.GetEducatorByUsernameWithCaseloadRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceGetEducatorByUsernameWithCaseload(conn, username)
		return err
	}); err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("educator not found")
	}

	caseManager := &caseloadStudentsModels.CaseManager{
		Educator: models.Educator{
			ID: rows[0].EducatorId,
			Person: sharedmodels.Person{
				GivenName:  rows[0].GivenName,
				ChosenName: rows[0].ChosenName,
				FamilyName: rows[0].FamilyName,
				Email:      rows[0].Email,
				Username:   rows[0].Username,
			},
		},
		Caseload: []studentModels.Student{}, // empty initially
	}
	for _, row := range rows {
		if row.StudentId != nil {
			student := studentModels.Student{
				ID: *row.StudentId,
				Person: sharedmodels.Person{
					GivenName:  *row.StudentGivenName,
					ChosenName: *row.StudentChosenName,
					FamilyName: *row.StudentFamilyName,
					Email:      *row.StudentEmail,
					Username:   *row.StudentUsername,
				},
				Grade:      sharedmodels.Grade(*row.Grade),
				HomeroomID: *row.HomeroomId,
			}
			caseManager.Caseload = append(caseManager.Caseload, student)
		}
	}

	return caseManager, nil
}

type ListOption func(*listConfig)

type listConfig struct {
	roleFilter *sharedmodels.EducatorRole
}

// returns only educators with the given role
func FilterByRole(role sharedmodels.EducatorRole) ListOption {
	return func(c *listConfig) { c.roleFilter = &role }
}

func (m *ReadModel) List(ctx context.Context, opts ...ListOption) ([]models.Educator, error) {
	cfg := &listConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.roleFilter != nil {
		return m.listByRole(ctx, *cfg.roleFilter)
	}
	return m.listAll(ctx)
}

func (m *ReadModel) listAll(ctx context.Context) ([]models.Educator, error) {
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
				Username:   rows[i].Username,
			},
		}
	}
	return educators, nil
}

func (m *ReadModel) listByRole(ctx context.Context, role sharedmodels.EducatorRole) ([]models.Educator, error) {
	var rows []dbsql.ListEducatorsByRoleRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListEducatorsByRole(conn, role.String())
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
				Username:   rows[i].Username,
			},
		}
	}
	return educators, nil
}

func (m *ReadModel) ListWithRoles(ctx context.Context) ([]models.Educator, error) {
	var rows []dbsql.ListEducatorsWithRolesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListEducatorsWithRoles(conn)
		return err
	}); err != nil {
		return nil, err
	}

	educatorMap := make(map[string]int)
	var educators []models.Educator

	for _, row := range rows {
		idx, ok := educatorMap[row.Id]
		if !ok {
			educator := models.Educator{
				ID: row.Id,
				Person: sharedmodels.Person{
					GivenName:  row.GivenName,
					ChosenName: row.ChosenName,
					FamilyName: row.FamilyName,
					Email:      row.Email,
					Username:   row.Username,
				},
				Roles: []sharedmodels.EducatorRole{},
			}
			educatorMap[row.Id] = len(educators)
			educators = append(educators, educator)
			idx = len(educators) - 1
		}

		if row.Role != nil {
			educators[idx].Roles = append(educators[idx].Roles, sharedmodels.EducatorRole(*row.Role))
		}
	}

	return educators, nil
}

// educator read model writer functions

func (m *ReadModel) Create(ctx context.Context, event EducatorCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		if err := dbsql.OnceCreateEducator(conn, dbsql.CreateEducatorParams{
			Id:                       event.ID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		}); err != nil {
			return err
		}
		for _, role := range event.Roles {
			if err := dbsql.OnceAddRoleToEducator(conn, dbsql.AddRoleToEducatorParams{
				EducatorId:               event.ID,
				Role:                     role,
				LastEventCommitPosition:  event.Position.Commit,
				LastEventPreparePosition: event.Position.Prepare,
				CreatedAt:                appdb.SQLTime(event.CreatedAt),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *ReadModel) Update(ctx context.Context, event EducatorUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateEducator(conn, dbsql.UpdateEducatorParams{
			Id:                       event.ID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
		})
	})
}

func (m *ReadModel) Archive(ctx context.Context, event EducatorArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveEducator(conn, dbsql.ArchiveEducatorParams{
			Id:                       event.ID,
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
		})
	})
}

func (m *ReadModel) Delete(ctx context.Context, event EducatorDeletedProjection) error {
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
