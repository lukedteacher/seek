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

	iepModels "seek/internal/features/ieps/models"
	"seek/internal/features/students/models"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type ReadModel struct {
	db *appdb.DB
}

func NewReadModel(db *appdb.DB) *ReadModel {
	return &ReadModel{db: db}
}

// student read model reader functions

func (m *ReadModel) GetByID(ctx context.Context, studentID string) (*models.Student, error) {
	var row *dbsql.GetStudentByIdRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudentById(conn, studentID)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("student not found")
	}

	student := &models.Student{
		ID:      row.Id,
		MARSSID: row.MarssId,
		Person: sharedmodels.Person{
			Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Pronouns:   parsePronouns(row.Pronouns),
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:      sharedmodels.Grade(row.Grade),
		HomeroomID: row.HomeroomId,
		PlanType:   sharedmodels.PlanType(row.PlanType),
	}

	return student, nil
}
func (m *ReadModel) GetByUsername(ctx context.Context, username string) (*models.Student, error) {
	var row *dbsql.GetStudentByUsernameRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		row, err = dbsql.OnceGetStudentByUsername(conn, username)
		return err
	}); err != nil {
		return nil, err
	}

	if row == nil {
		return nil, fmt.Errorf("student not found")
	}

	student := &models.Student{
		ID:      row.Id,
		MARSSID: row.MarssId,
		Person: sharedmodels.Person{
			Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Pronouns:   parsePronouns(row.Pronouns),
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:      sharedmodels.Grade(row.Grade),
		HomeroomID: row.HomeroomId,
		PlanType:   sharedmodels.PlanType(row.PlanType),
	}

	return student, nil
}

type ListOption func(*listConfig)

type listConfig struct {
	withCaseManager bool
	withServices    bool
	withGradeFilter []int
	withPlanFilter  []int
	withSort        struct {
		column    string
		direction string
	}
}

func WithCaseManager() ListOption {
	return func(c *listConfig) {
		c.withCaseManager = true
	}
}

func WithServices() ListOption {
	return func(c *listConfig) {
		c.withServices = true
	}
}

func WithSort(col, dir string) ListOption {
	return func(c *listConfig) {
		c.withSort.column = col
		c.withSort.direction = dir
	}
}

func WithGradeFilter(grades []int) ListOption {
	return func(c *listConfig) {
		c.withGradeFilter = grades
	}
}

func WithPlanFilter(plans []int) ListOption {
	return func(c *listConfig) {
		c.withPlanFilter = plans
	}
}

func (m *ReadModel) List(ctx context.Context, opts ...ListOption) ([]models.Student, error) {
	cfg := &listConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.withCaseManager {
		// TODO later?
	}
	// TODO fix this
	if cfg.withSort.column != "" && cfg.withSort.direction != "" {
		return m.listAllWithSorting(ctx, cfg.withSort.column, cfg.withSort.direction, cfg.withGradeFilter, cfg.withPlanFilter)
	}
	if len(cfg.withGradeFilter) > 0 && len(cfg.withGradeFilter) < 9 {
		return m.listByGrade(ctx, cfg.withGradeFilter)
	}
	return m.listAll(ctx)
}

func (m *ReadModel) listAllWithSorting(
	ctx context.Context,
	sortBy,
	sortDir string,
	grades []int,
	plans []int,
) ([]models.Student, error) {
	allowedColumns := map[string]bool{
		"marss_id":     true,
		"birthdate":    true,
		"given_name":   true,
		"chosen_name":  true,
		"family_name":  true,
		"email":        true,
		"grade":        true,
		"homeroom_id":  true,
		"plan_type":    true,
		"case_manager": true,
		"created_at":   true,
		"updated_at":   true,
	}
	allowedDirs := map[string]bool{"ASC": true, "DESC": true}

	if !allowedColumns[sortBy] {
		sortBy = "family_name"
	}
	if !allowedDirs[sortDir] {
		sortDir = "DESC"
	}

	// Build WHERE clause
	where := "archived_at IS NULL"
	args := []any{}
	if len(grades) > 0 {
		placeholders := strings.Repeat("?,", len(grades))
		placeholders = placeholders[:len(placeholders)-1] // trim trailing comma
		where += " AND grade IN (" + placeholders + ")"
		for _, g := range grades {
			args = append(args, g)
		}
	}
	if len(plans) > 0 {
		placeholders := strings.Repeat("?,", len(plans))
		placeholders = placeholders[:len(placeholders)-1] // trim trailing comma
		where += " AND plan_type IN (" + placeholders + ")"
		for _, p := range plans {
			args = append(args, p)
		}
	}

	query := fmt.Sprintf(`
			SELECT 
				id, marss_id, birthdate, given_name, chosen_name, family_name, pronouns,
				email, username, grade, homeroom_id, plan_type,
				created_at, updated_at
			FROM students
			WHERE %s
			ORDER BY %s %s, family_name ASC, given_name ASC
    `, where, sortBy, sortDir)

	var students []models.Student
	err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, query, &sqlitex.ExecOptions{
			Args: args,
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var student models.Student
				student.ID = stmt.ColumnText(0)
				student.MARSSID = stmt.ColumnText(1)
				student.Birthdate = sharedmodels.DateOnly(parseDBTime(stmt.ColumnText(2)))
				student.GivenName = stmt.ColumnText(3)
				student.ChosenName = stmt.ColumnText(4)
				student.FamilyName = stmt.ColumnText(5)
				student.Pronouns = parsePronouns(stmt.ColumnText(6))
				student.Email = stmt.ColumnText(7)
				student.Username = stmt.ColumnText(8)
				student.Grade = sharedmodels.Grade(stmt.ColumnInt64(9))
				student.HomeroomID = stmt.ColumnText(10)
				student.PlanType = sharedmodels.PlanType(stmt.ColumnInt(11))
				student.CreatedAt = parseDBTime(stmt.ColumnText(12))
				student.UpdatedAt = parseDBTime(stmt.ColumnText(13))
				students = append(students, student)
				return nil
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (m *ReadModel) listAll(ctx context.Context) ([]models.Student, error) {
	var rows []dbsql.ListStudentsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudents(conn)
		return err
	}); err != nil {
		return nil, err
	}

	students := make([]models.Student, len(rows))
	for i := range rows {
		row := rows[i]
		students[i] = models.Student{
			ID:      row.Id,
			MARSSID: row.MarssId,
			Person: sharedmodels.Person{
				Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
				GivenName:  row.GivenName,
				ChosenName: row.ChosenName,
				FamilyName: row.FamilyName,
				Pronouns:   parsePronouns(row.Pronouns),
				Email:      row.Email,
				Username:   row.Username,
			},
			Grade:      sharedmodels.Grade(row.Grade),
			HomeroomID: row.HomeroomId,
			PlanType:   sharedmodels.PlanType(row.PlanType),
		}
	}
	return students, nil
}

func (m *ReadModel) listByGrade(ctx context.Context, grades []int) ([]models.Student, error) {
	var rows []dbsql.ListStudentsByGradeRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		gradeInt64 := make([]int64, len(grades))
		for i, g := range grades {
			gradeInt64[i] = int64(g)
		}
		var err error
		rows, err = dbsql.OnceListStudentsByGrade(conn, gradeInt64)
		return err
	}); err != nil {
		return nil, err
	}

	students := make([]models.Student, len(rows))
	for i := range rows {
		row := rows[i]
		students[i] = models.Student{
			ID:      row.Id,
			MARSSID: row.MarssId,
			Person: sharedmodels.Person{
				Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
				GivenName:  row.GivenName,
				ChosenName: row.ChosenName,
				FamilyName: row.FamilyName,
				Pronouns:   parsePronouns(row.Pronouns),
				Email:      row.Email,
				Username:   row.Username,
			},
			Grade:      sharedmodels.Grade(row.Grade),
			HomeroomID: row.HomeroomId,
			PlanType:   sharedmodels.PlanType(row.PlanType),
		}
	}
	return students, nil
}

func (m *ReadModel) ListWithIEPs(ctx context.Context) ([]models.StudentWithIEP, error) {
	var rows []dbsql.ListOnlyStudentsWithIepsRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListOnlyStudentsWithIeps(conn)
		return err
	}); err != nil {
		return nil, err
	}
	// map student ID -> index in the final slice
	studentMap := make(map[string]int)
	var students []models.StudentWithIEP

	for _, row := range rows {
		idx, ok := studentMap[row.StudentId]
		if !ok {
			// create a new student entry
			student := models.StudentWithIEP{
				ID:      row.StudentId,
				MARSSID: row.MarssId,
				Person: sharedmodels.Person{
					Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
					GivenName:  row.GivenName,
					ChosenName: row.ChosenName,
					FamilyName: row.FamilyName,
					Pronouns:   parsePronouns(row.Pronouns),
					Email:      row.Email,
					Username:   row.Username,
				},
				Grade:      sharedmodels.Grade(row.Grade),
				HomeroomID: row.HomeroomId,
				PlanType:   sharedmodels.PlanType(row.PlanType),
				CreatedAt:  parseDBTime(row.CreatedAt),
				UpdatedAt:  parseDBTime(row.UpdatedAt),
			}
			studentMap[row.StudentId] = len(students)
			students = append(students, student)
			idx = len(students) - 1
		}
		// append the IEP to the existing student
		iep := iepModels.IEP{
			ID:          row.IepId,
			StudentID:   row.StudentId,
			StartDate:   sharedmodels.DateOnly(parseDBTime(row.StartDate)),
			EndDate:     sharedmodels.DateOnly(parseDBTime(row.EndDate)),
			AmendedDate: sharedmodels.DateOnly(parseDBTime(row.AmendedDate)),
			CreatedAt:   parseDBTime(row.IepCreatedAt),
			UpdatedAt:   parseDBTime(row.IepUpdatedAt),
		}
		students[idx].IEP = iep
	}

	return students, nil
}

func (m *ReadModel) ListByServiceType(ctx context.Context, serviceType string) ([]models.Student, error) {
	var rows []dbsql.ListStudentsByServiceTypeRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentsByServiceType(conn, serviceType)
		return err
	}); err != nil {
		return nil, err
	}

	students := make([]models.Student, len(rows))
	for i := range rows {
		row := rows[i]
		students[i] = models.Student{
			ID:      row.Id,
			MARSSID: row.MarssId,
			Person: sharedmodels.Person{
				Birthdate:  sharedmodels.DateOnly(parseDBTime(row.Birthdate)),
				GivenName:  row.GivenName,
				ChosenName: row.ChosenName,
				FamilyName: row.FamilyName,
				Pronouns:   parsePronouns(row.Pronouns),
				Email:      row.Email,
				Username:   row.Username,
			},
			Grade:      sharedmodels.Grade(row.Grade),
			HomeroomID: row.HomeroomId,
			PlanType:   sharedmodels.PlanType(row.PlanType),
		}
	}
	return students, nil
}

// student read model writer functions

func (m *ReadModel) Create(ctx context.Context, event StudentCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateStudent(conn, dbsql.CreateStudentParams{
			Id:                       event.ID,
			MarssId:                  event.MARSSID,
			Birthdate:                event.Birthdate,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Pronouns:                 event.Pronouns,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			HomeroomId:               event.HomeroomID,
			PlanType:                 int64(event.PlanType),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) Update(ctx context.Context, event StudentUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateStudent(conn, dbsql.UpdateStudentParams{
			Id:                       event.ID,
			MarssId:                  event.MARSSID,
			Birthdate:                event.Birthdate,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Pronouns:                 event.Pronouns,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			HomeroomId:               event.HomeroomID,
			PlanType:                 int64(event.PlanType),
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
		})
	})
}

func (m *ReadModel) Archive(ctx context.Context, event StudentArchivedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceArchiveStudent(conn, dbsql.ArchiveStudentParams{
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			ArchivedAt:               appdb.SQLTime(event.ArchivedAt),
			Id:                       event.StudentID,
		})
	})
}

func (m *ReadModel) Delete(ctx context.Context, event StudentDeletedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceDeleteStudent(conn, event.StudentID)
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

func parsePronouns(s string) []sharedmodels.Pronoun {
	var p []sharedmodels.Pronoun
	if s != "" {
		_ = json.Unmarshal([]byte(s), &p)
	}
	return p
}
