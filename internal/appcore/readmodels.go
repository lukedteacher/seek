package appcore

import (
	"seek/internal/appdb"
	educatorEvents "seek/internal/features/educators/events"
	educatorPeriodEvents "seek/internal/features/educators_periods/events"
	iepServiceEvents "seek/internal/features/iepservices/events"
	periodEvents "seek/internal/features/periods/events"
	profileEvents "seek/internal/features/profiles/events"
	studentEvents "seek/internal/features/students/events"
	studentPeriodEvents "seek/internal/features/students_periods/events"
)

type ReadModelContainer struct {
	Educators       *educatorEvents.ReadModel
	EducatorPeriods *educatorPeriodEvents.ReadModel
	IEPServices     *iepServiceEvents.ReadModel
	Periods         *periodEvents.ReadModel
	Profiles        *profileEvents.ReadModel
	Students        *studentEvents.ReadModel
	StudentPeriods  *studentPeriodEvents.ReadModel
}

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
