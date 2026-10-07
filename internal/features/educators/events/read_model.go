package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"
	caseloadStudentsModels "seek/internal/features/caseload_students/models"
	"seek/internal/features/educators/dto"
	"seek/internal/features/educators/models"
	periodModels "seek/internal/features/periods/models"
	studentModels "seek/internal/features/students/models"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
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
			Pronouns:   parsePronouns(row.Pronouns),
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
			Pronouns:   parsePronouns(row.Pronouns),
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
			Pronouns:   parsePronouns(rows[0].Pronouns),
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
				Pronouns:   parsePronouns(rows[0].Pronouns),
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

func (m *ReadModel) GetWithPeriods(ctx context.Context, username string) (*dto.EducatorWithPeriods, error) {
	// 1. execute the query inside a read transaction.
	var row *dbsql.GetEducatorWithPeriodsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetEducatorWithPeriods(conn, username)
		return err
	}); err != nil {
		return nil, err
	}

	// 2. if no educator found, return nil.
	if row.Id == "" {
		return nil, nil
	}

	// 3. build the base educator from the scalar columns.
	educator := models.Educator{
		ID: row.Id,
		Person: sharedmodels.Person{
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Pronouns:   parsePronouns(row.Pronouns),
			Email:      row.Email,
			Username:   row.Username,
		},
	}

	// 4. unmarshal periods directly into the domain slice.
	var periods []periodModels.Period
	if row.PeriodsJson != "" && row.PeriodsJson != "[]" {
		if err := json.Unmarshal([]byte(row.PeriodsJson), &periods); err != nil {
			return nil, fmt.Errorf("unmarshal periods: %w", err)
		}
	}

	// 5. unmarshal students directly into the map keyed by period id.
	studentsMap := map[string][]studentModels.Student{}
	if row.StudentsJson != "" && row.StudentsJson != "{}" {
		if err := json.Unmarshal([]byte(row.StudentsJson), &studentsMap); err != nil {
			return nil, fmt.Errorf("unmarshal students: %w", err)
		}
	}

	// 6. EndTime is computed, not stored — fill it in.
	for i := range periods {
		periods[i].EndTime = periods[i].StartTime.Add(periods[i].Duration)
	}

	// 7. return the assembled aggregate.
	return &dto.EducatorWithPeriods{
		Educator:          educator,
		Periods:           periods,
		PeriodStudentsMap: studentsMap,
	}, nil
}

type ListOption func(*listConfig)

type listConfig struct {
	withRoleFilter   []string
	withSearchFilter string
}

func WithSearchFilter(search string) ListOption {
	return func(c *listConfig) {
		c.withSearchFilter = search
	}
}

func WithRoleFilter(roles []string) ListOption {
	return func(c *listConfig) {
		c.withRoleFilter = roles
	}
}

func (m *ReadModel) List(ctx context.Context, opts ...ListOption) ([]models.Educator, error) {
	cfg := &listConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return m.listWithFilters(ctx, cfg)
}

func (m *ReadModel) listWithFilters(
	ctx context.Context,
	cfg *listConfig,
) ([]models.Educator, error) {
	where := "e.archived_at IS NULL"
	args := []any{}
	if cfg.withSearchFilter != "" {
		searchTerm := "%" + cfg.withSearchFilter + "%"
		where += " AND ("
		cols := []string{"e.given_name", "e.family_name", "e.chosen_name"}
		for i, col := range cols {
			if i > 0 {
				where += " OR "
			}
			where += col + " LIKE ?"
			args = append(args, searchTerm)
		}
		where += ")"
	}

	if cfg.withRoleFilter != nil {
		if len(cfg.withRoleFilter) == 0 {
			return []models.Educator{}, nil
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cfg.withRoleFilter)), ",")
		where += " AND EXISTS (SELECT 1 FROM educator_roles er " +
			"WHERE er.educator_id = e.id AND er.archived_at IS NULL " +
			"AND er.role IN (" + placeholders + "))"
		for _, role := range cfg.withRoleFilter {
			args = append(args, role)
		}
	}

	query := fmt.Sprintf(`
		SELECT
			e.id, e.given_name, e.chosen_name, e.family_name, e.pronouns,
			e.email, e.username,
			e.created_at, e.updated_at,
			COALESCE(
				(SELECT json_group_array(er2.role)
				 FROM educator_roles er2
				 WHERE er2.educator_id = e.id AND er2.archived_at IS NULL),
				'[]'
			) AS roles_json
		FROM educators e
		WHERE %s
		ORDER BY e.family_name ASC, e.given_name ASC
	`, where)

	var educators []models.Educator
	err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, query, &sqlitex.ExecOptions{
			Args: args,
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var e models.Educator
				e.ID = stmt.ColumnText(0)
				e.GivenName = stmt.ColumnText(1)
				e.ChosenName = stmt.ColumnText(2)
				e.FamilyName = stmt.ColumnText(3)
				e.Pronouns = parsePronouns(stmt.ColumnText(4))
				e.Email = stmt.ColumnText(5)
				e.Username = stmt.ColumnText(6)
				e.CreatedAt = parseDBTime(stmt.ColumnText(7))
				e.UpdatedAt = parseDBTime(stmt.ColumnText(8))

				var roles []sharedmodels.EducatorRole
				if err := json.Unmarshal([]byte(stmt.ColumnText(9)), &roles); err != nil {
					return err
				}
				e.Roles = roles

				educators = append(educators, e)
				return nil
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return educators, nil
}

func (m *ReadModel) ListByIDs(ctx context.Context, educatorIDs []string) ([]models.Educator, error) {
	var rows []dbsql.ListEducatorsByIdsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListEducatorsByIds(conn, educatorIDs)
		return err
	}); err != nil {
		return nil, err
	}

	educators := make([]models.Educator, len(rows))
	for i := range rows {
		row := rows[i]
		educators[i] = models.Educator{
			ID: row.Id,
			Person: sharedmodels.Person{
				GivenName:  row.GivenName,
				ChosenName: row.ChosenName,
				FamilyName: row.FamilyName,
				Pronouns:   parsePronouns(row.Pronouns),
				Email:      row.Email,
				Username:   row.Username,
			},
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
			Pronouns:                 event.Pronouns,
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
		if err := dbsql.OnceUpdateEducator(conn, dbsql.UpdateEducatorParams{
			Id:                       event.ID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Pronouns:                 event.Pronouns,
			Email:                    event.Email,
			Username:                 event.Username,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
		}); err != nil {
			return err
		}
		var previousRoles []string
		if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
			var err error
			previousRoles, err = dbsql.OnceGetRolesForEducator(conn, event.ID)
			return err
		}); err != nil {
			return err
		}
		previousRoleMap := make(map[string]struct{}, len(previousRoles))
		for _, role := range previousRoles {
			previousRoleMap[role] = struct{}{}
		}
		proposedRoleMap := make(map[string]struct{}, len(event.Roles))
		for _, role := range event.Roles {
			proposedRoleMap[role] = struct{}{}
			_, ok := previousRoleMap[role]
			if !ok {
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
		}
		for _, role := range previousRoles {
			_, ok := proposedRoleMap[role]
			if !ok {
				if err := dbsql.OnceRemoveRoleFromEducator(conn, dbsql.RemoveRoleFromEducatorParams{
					EducatorId: event.ID,
					Role:       role,
				}); err != nil {
					return err
				}
			}
		}
		return nil
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

func parseDBTimeOnly(value string) sharedmodels.TimeOnly {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return sharedmodels.TimeOnly(parsed)
		}
	}
	return sharedmodels.TimeOnly{}
}

func parsePronouns(s string) []sharedmodels.Pronoun {
	var p []sharedmodels.Pronoun
	if s != "" {
		_ = json.Unmarshal([]byte(s), &p)
	}
	return p
}
