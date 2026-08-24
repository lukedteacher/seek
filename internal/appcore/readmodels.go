package appcore

import (
	"context"
	"seek/internal/appdb"
	"seek/internal/dbsql"
	caseloadStudentsEvents "seek/internal/features/caseload_students/events"
	educatorEvents "seek/internal/features/educators/events"
	educatorPeriodEvents "seek/internal/features/educators_periods/events"
	iepEvents "seek/internal/features/ieps/events"
	iepServiceEvents "seek/internal/features/iepservices/events"
	periodEvents "seek/internal/features/periods/events"
	profileEvents "seek/internal/features/profiles/events"
	studentEvents "seek/internal/features/students/events"
	studentPeriodEvents "seek/internal/features/students_periods/events"

	"zombiezen.com/go/sqlite"
)

type ReadModelContainer struct {
	CaseloadStudents *caseloadStudentsEvents.ReadModel
	Educators        *educatorEvents.ReadModel
	EducatorPeriods  *educatorPeriodEvents.ReadModel
	IEPServices      *iepServiceEvents.ReadModel
	Periods          *periodEvents.ReadModel
	Profiles         *profileEvents.ReadModel
	Students         *studentEvents.ReadModel
	IEPs             *iepEvents.ReadModel
	StudentPeriods   *studentPeriodEvents.ReadModel
}

func NewReadModelContainer(db *appdb.DB) *ReadModelContainer {
	return &ReadModelContainer{
		CaseloadStudents: caseloadStudentsEvents.NewReadModel(db),
		Educators:        educatorEvents.NewReadModel(db),
		EducatorPeriods:  educatorPeriodEvents.NewReadModel(db),
		IEPServices:      iepServiceEvents.NewReadModel(db),
		Periods:          periodEvents.NewReadModel(db),
		Profiles:         profileEvents.NewReadModel(db),
		Students:         studentEvents.NewReadModel(db),
		IEPs:             iepEvents.NewReadModel(db),
		StudentPeriods:   studentPeriodEvents.NewReadModel(db),
	}
}

func (c *ReadModelContainer) Reset(ctx context.Context, db *appdb.DB) error {
	return db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		if err := dbsql.OnceResetReadModelAuthSessions(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelAuthAccounts(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelProfiles(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelIepservices(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelPeriods(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelStudents(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelPeriodsStudents(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetReadModelAuthUsers(conn); err != nil {
			return err
		}
		if err := dbsql.OnceResetEventHandlerCheckpoints(conn); err != nil {
			return err
		}
		return nil
	})
}
