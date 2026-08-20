package events

import (
	"context"
	"fmt"
	"time"

	"seek/internal/appdb"
	"seek/internal/dbsql"
	"seek/internal/features/_shared/sharedmodels"

	// smodels "seek/internal/features/iepservices/models"
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
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:         sharedmodels.Grade(row.Grade),
		Homeroom:      row.Homeroom,
		CaseManagerID: row.CaseManager,
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
			GivenName:  row.GivenName,
			ChosenName: row.ChosenName,
			FamilyName: row.FamilyName,
			Email:      row.Email,
			Username:   row.Username,
		},
		Grade:         sharedmodels.Grade(row.Grade),
		Homeroom:      row.Homeroom,
		CaseManagerID: row.CaseManager,
	}

	return student, nil
}

type ListOption func(*listConfig)

type listConfig struct {
	withCaseManager bool
	withServices    bool
	withGradeFilter []int
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

func (m *ReadModel) List(ctx context.Context, opts ...ListOption) ([]models.Student, error) {
	cfg := &listConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.withCaseManager {
		// TODO later?
	}
	if cfg.withSort.column != "" && cfg.withSort.direction != "" {
		return m.listAllWithSorting(ctx, cfg.withSort.column, cfg.withSort.direction)
	}
	if len(cfg.withGradeFilter) > 0 && len(cfg.withGradeFilter) < 9 {
		return m.listByGrade(ctx, cfg.withGradeFilter)
	}
	if cfg.withServices {
		return m.listWithServices(ctx)
	}
	return m.listAll(ctx)
}

func (m *ReadModel) listAllWithSorting(
	ctx context.Context,
	sortBy,
	sortDir string,
) ([]models.Student, error) {
	allowedColumns := map[string]bool{
		"marss_id":     true,
		"given_name":   true,
		"chosen_name":  true,
		"family_name":  true,
		"email":        true,
		"grade":        true,
		"homeroom":     true,
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

	query := fmt.Sprintf(`
        SELECT 
            id, marss_id, given_name, chosen_name, family_name,
            email, username, grade, homeroom, case_manager,
            created_at, updated_at
        FROM students
        WHERE archived_at IS NULL
        ORDER BY %s %s, family_name ASC, given_name ASC
    `, sortBy, sortDir)

	var students []models.Student
	err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, query, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				var student models.Student
				student.ID = stmt.ColumnText(0)
				student.MARSSID = stmt.ColumnText(1)
				student.GivenName = stmt.ColumnText(2)
				student.ChosenName = stmt.ColumnText(3)
				student.FamilyName = stmt.ColumnText(4)
				student.Email = stmt.ColumnText(5)
				student.Username = stmt.ColumnText(6)
				student.Grade = sharedmodels.Grade(stmt.ColumnInt64(7))
				student.Homeroom = stmt.ColumnText(8)
				student.CaseManagerID = stmt.ColumnText(9)
				student.CreatedAt = parseDBTime(stmt.ColumnText(10))
				student.UpdatedAt = parseDBTime(stmt.ColumnText(11))
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
		students[i] = models.Student{
			ID:      rows[i].Id,
			MARSSID: rows[i].MarssId,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
				Username:   rows[i].Username,
			},
			Grade:         sharedmodels.Grade(rows[i].Grade),
			Homeroom:      rows[i].Homeroom,
			CaseManagerID: rows[i].CaseManager,
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
		students[i] = models.Student{
			ID:      rows[i].Id,
			MARSSID: rows[i].MarssId,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
				Username:   rows[i].Username,
			},
			Grade:         sharedmodels.Grade(rows[i].Grade),
			Homeroom:      rows[i].Homeroom,
			CaseManagerID: rows[i].CaseManager,
		}
	}
	return students, nil
}

func (m *ReadModel) listWithServices(ctx context.Context) ([]models.Student, error) {
	var rows []dbsql.ListStudentsWithIepservicesRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentsWithIepservices(conn)
		return err
	}); err != nil {
		return nil, err
	}
	// map student ID -> index in the final slice
	studentMap := make(map[string]int)
	var students []models.Student

	for _, row := range rows {
		_, ok := studentMap[row.StudentId]
		if !ok {
			// create a new student entry
			student := models.Student{
				ID:      row.StudentId,
				MARSSID: row.MarssId,
				Person: sharedmodels.Person{
					GivenName:  row.GivenName,
					ChosenName: row.ChosenName,
					FamilyName: row.FamilyName,
					Email:      row.Email,
					Username:   row.Username,
				},
				Grade:         sharedmodels.Grade(row.Grade),
				Homeroom:      row.Homeroom,
				CaseManagerID: row.CaseManager,
				CreatedAt:     parseDBTime(row.CreatedAt),
				UpdatedAt:     parseDBTime(row.UpdatedAt),
			}
			studentMap[row.StudentId] = len(students)
			students = append(students, student)
			// idx = len(students) - 1
		}

		// if row.ServiceId != nil {
		// 	// append the service to the existing student
		// 	service := smodels.IEPService{
		// 		ID:              *row.ServiceId,
		// 		StudentID:       row.StudentId,
		// 		ServiceType:     sharedmodels.ServiceType(*row.ServiceType),
		// 		IndirectMinutes: int(*row.IndirectMinutes),
		// 		DirectMinutes:   int(*row.DirectMinutes),
		// 		FrequencyCount:  int(*row.FrequencyCount),
		// 		FrequencyType:   *row.FrequencyType,
		// 		Location:        *row.Location,
		// 		StartDate:       sharedmodels.DateOnly(parseDBTime(*row.StartDate)),
		// 		EndDate:         sharedmodels.DateOnly(parseDBTime(*row.EndDate)),
		// 		Provider:        *row.Provider,
		// 		CreatedAt:       parseDBTime(*row.ServiceCreatedAt),
		// 		UpdatedAt:       parseDBTime(*row.ServiceUpdatedAt),
		// 	}
		// 	students[idx].Services = append(students[idx].Services, service)
		// }
	}

	return students, nil
}

func (m *ReadModel) ListByIEPServiceType(ctx context.Context, serviceType string) ([]models.Student, error) {
	var rows []dbsql.ListStudentsByIepserviceTypeRes
	if err := m.db.ReadTX(ctx, func(conn *sqlite.Conn) error {
		var err error
		rows, err = dbsql.OnceListStudentsByIepserviceType(conn, serviceType)
		return err
	}); err != nil {
		return nil, err
	}

	students := make([]models.Student, len(rows))
	for i := range rows {
		students[i] = models.Student{
			ID:      rows[i].Id,
			MARSSID: rows[i].MarssId,
			Person: sharedmodels.Person{
				GivenName:  rows[i].GivenName,
				ChosenName: rows[i].ChosenName,
				FamilyName: rows[i].FamilyName,
				Email:      rows[i].Email,
				Username:   rows[i].Username,
			},
			Grade:         sharedmodels.Grade(rows[i].Grade),
			Homeroom:      rows[i].Homeroom,
			CaseManagerID: rows[i].CaseManager,
		}
	}
	return students, nil
}

// student read model writer functions

func (m *ReadModel) Create(ctx context.Context, event StudentCreatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceCreateStudent(conn, dbsql.CreateStudentParams{
			Id:                       event.StudentID,
			MarssId:                  event.MARSSID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			CreatedAt:                appdb.SQLTime(event.CreatedAt),
		})
	})
}

func (m *ReadModel) Update(ctx context.Context, event StudentUpdatedProjection) error {
	return m.db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return dbsql.OnceUpdateStudent(conn, dbsql.UpdateStudentParams{
			MarssId:                  event.MARSSID,
			GivenName:                event.GivenName,
			ChosenName:               event.ChosenName,
			FamilyName:               event.FamilyName,
			Email:                    event.Email,
			Username:                 event.Username,
			Grade:                    int64(event.Grade),
			Homeroom:                 event.Homeroom,
			CaseManager:              event.CaseManager,
			LastEventCommitPosition:  event.Position.Commit,
			LastEventPreparePosition: event.Position.Prepare,
			UpdatedAt:                appdb.SQLTime(event.UpdatedAt),
			Id:                       event.StudentID,
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
