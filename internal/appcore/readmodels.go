package appcore

import (
	"context"
	"seek/internal/appdb"
	"seek/internal/dbsql"
	educatorEvents "seek/internal/features/educators/events"
	educatorPeriodEvents "seek/internal/features/educators_periods/events"
	iepServiceEvents "seek/internal/features/iepservices/events"
	periodEvents "seek/internal/features/periods/events"
	profileEvents "seek/internal/features/profiles/events"
	studentEvents "seek/internal/features/students/events"
	studentPeriodEvents "seek/internal/features/students_periods/events"

	"zombiezen.com/go/sqlite"
)

// ReadModelContainer holds all read model instances used by the application.
// It provides a single point of access to the various read models, reducing
// the number of fields needed in handlers and wiring code.
type ReadModelContainer struct {
	// Educators provides access to educator data.
	Educators *educatorEvents.ReadModel
	// EducatorPeriods manages the many-to-many relationship between educators and periods.
	EducatorPeriods *educatorPeriodEvents.ReadModel
	// IEPServices handles IEP services data for students.
	IEPServices *iepServiceEvents.ReadModel
	// Periods holds period schedule data.
	Periods *periodEvents.ReadModel
	// Profiles stores user profile information.
	Profiles *profileEvents.ReadModel
	// Students provides student data.
	Students *studentEvents.ReadModel
	// StudentPeriods manages the many-to-many relationship between students and periods.
	StudentPeriods *studentPeriodEvents.ReadModel
}

// NewReadModelContainer creates and initializes all read models using the provided
// database connection. It returns a fully populated container ready for use.
func NewReadModelContainer(db *appdb.DB) *ReadModelContainer {
	return &ReadModelContainer{
		Educators:       educatorEvents.NewReadModel(db),
		EducatorPeriods: educatorPeriodEvents.NewReadModel(db),
		IEPServices:     iepServiceEvents.NewReadModel(db),
		Periods:         periodEvents.NewReadModel(db),
		Profiles:        profileEvents.NewReadModel(db),
		Students:        studentEvents.NewReadModel(db),
		StudentPeriods:  studentPeriodEvents.NewReadModel(db),
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
		if err := dbsql.OnceResetReadModelAuthVerifications(conn); err != nil {
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
