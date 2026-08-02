package seed

import (
	"context"
	"log/slog"
	"time"

	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	ee "seek/internal/features/educators/events"
	epe "seek/internal/features/educators_periods/events"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
	spe "seek/internal/features/students_periods/events"
)

func SeedData(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	piiKeys *auth.SubjectPiiKeyStore,
	logger *slog.Logger,
) error {
	err := seedUsers(ctx, saver, retriever, piiKeys)
	if err != nil {
		logger.ErrorContext(ctx, "seed data user", "err", err)
		return err
	}
	studentIDs, err := seedStudents(ctx, saver)
	if err != nil {
		logger.ErrorContext(ctx, "seed students", "err", err)
		return err
	}
	periodIDs, err := seedPeriods(ctx, saver)
	if err != nil {
		logger.ErrorContext(ctx, "seed periods", "err", err)
		return err
	}
	educatorIDs, err := seedEducators(ctx, saver)
	if err != nil {
		logger.ErrorContext(ctx, "seed educators", "err", err)
		return err
	}
	if len(educatorIDs) > 0 && len(periodIDs) > 0 {
		for i, eduID := range educatorIDs {
			periodIdx := i % len(periodIDs)
			if err := seedEducatorPeriods(ctx, saver, retriever, eduID, periodIDs[periodIdx]); err != nil {
				logger.ErrorContext(ctx, "seed educator period", "err", err, "educator", eduID, "period", periodIDs[periodIdx])
				return err
			}
		}
	}
	if len(studentIDs) > 0 && len(periodIDs) > 1 {
		for _, stuID := range studentIDs {
			if err := seedStudentPeriods(ctx, saver, retriever, stuID, periodIDs[0]); err != nil {
				logger.ErrorContext(ctx, "seed student period 1", "err", err, "student", stuID)
				return err
			}
			if err := seedStudentPeriods(ctx, saver, retriever, stuID, periodIDs[1]); err != nil {
				logger.ErrorContext(ctx, "seed student period 2", "err", err, "student", stuID)
				return err
			}
		}
	}
	return nil
}

func seedUsers(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	piiKeys *auth.SubjectPiiKeyStore,
) error {
	_, err := auth.RegisterUserCommandHandler(ctx, auth.RegisterUserCommand{
		Email:    "luke@lukeout.world",
		Password: "fart123",
	}, saver, retriever, piiKeys)
	if err != nil {
		return err
	}
	return nil
}

func seedStudents(ctx context.Context, saver eventstore.Saver) ([]string, error) {
	data := []struct {
		GivenName   string
		ChosenName  string
		FamilyName  string
		Email       string
		Grade       int
		Homeroom    string
		CaseManager string
	}{
		{"Emma", "Em", "Johnson", "emma.j@schoolofnorthernlights.org", 5, "Ms. Smith", "Mr. Davis"},
		{"Liam", "", "Garcia", "liam.g@schoolofnorthernlights.org", 6, "Mr. Brown", "Ms. Wilson"},
		{"Olivia", "Liv", "Martinez", "olivia.m@schoolofnorthernlights.org", 2, "Mrs. Lee", ""},
		{"Noah", "", "Rodriguez", "noah.r@schoolofnorthernlights.org", 1, "Mr. Jones", "Ms. Taylor"},
		{"Ava", "Aves", "Williams", "ava.w@schoolofnorthernlights.org", 4, "Mrs. Clark", "Mr. Harris"},
		{"Mason", "Mace", "Brown", "mason.b@schoolofnorthernlights.org", 8, "Ms. White", ""},
		{"Sophia", "Soph", "Jones", "sophia.j@schoolofnorthernlights.org", 0, "Mr. Anderson", "Ms. Thomas"},
		{"Logan", "", "Miller", "logan.m@schoolofnorthernlights.org", 7, "Mrs. Martinez", "Mr. Garcia"},
		{"Mia", "", "Davis", "mia.d@schoolofnorthernlights.org", 5, "Ms. Smith", "Ms. Wilson"},
		{"Ethan", "E", "Moore", "ethan.m@schoolofnorthernlights.org", 6, "Mr. Brown", ""},
	}

	ids := make([]string, 0, len(data))
	for _, d := range data {
		cmd := se.CreateStudentCommand{
			GivenName:   d.GivenName,
			ChosenName:  d.ChosenName,
			FamilyName:  d.FamilyName,
			Email:       d.Email,
			Grade:       d.Grade,
			Homeroom:    d.Homeroom,
			CaseManager: d.CaseManager,
		}
		res, err := se.CreateStudentCommandHandler(ctx, cmd, saver)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.EventID)
	}
	return ids, nil
}

func seedPeriods(ctx context.Context, saver eventstore.Saver) ([]string, error) {
	data := []struct {
		Title       string
		ServiceType sharedmodels.ServiceType
		StartTime   sharedmodels.TimeOnly
		Duration    int
		DaysBitmask sharedmodels.DaysBitmask
	}{
		{"SEL", "SEL", parseTimeOnly("08:00"), 20, 1},
		{"reading", "reading", parseTimeOnly("08:55"), 30, 5},
		{"writing", "writing", parseTimeOnly("09:50"), 30, 15},
		{"SEL", "SEL", parseTimeOnly("10:50"), 20, 8},
		{"executive functioning", "EF", parseTimeOnly("11:45"), 15, 19},
		{"OT", "OT", parseTimeOnly("13:00"), 30, 18},
		{"speech", "speech", parseTimeOnly("13:40"), 20, 22},
		{"speech", "speech", parseTimeOnly("14:35"), 20, 29},
		{"math (push-in)", "math", parseTimeOnly("13:30"), 60, 10},
		{"reading", "reading", parseTimeOnly("14:20"), 30, 17},
	}

	ids := make([]string, 0, len(data))
	for _, d := range data {
		cmd := pe.CreatePeriodCommand{
			Title:       d.Title,
			ServiceType: d.ServiceType,
			StartTime:   d.StartTime,
			Duration:    d.Duration,
			DaysBitmask: d.DaysBitmask,
		}
		res, err := pe.CreatePeriodCommandHandler(ctx, cmd, saver)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.EventID)
	}
	return ids, nil
}

func seedEducators(ctx context.Context, saver eventstore.Saver) ([]string, error) {
	data := []struct {
		GivenName  string
		ChosenName string
		FamilyName string
		Email      string
		Role       string
	}{
		{"James", "Jim", "Peterson", "james.p@schoolofnorthernlights.org", "teacher"},
		{"Sarah", "Sally", "Chen", "sarah.c@schoolofnorthernlights.org", "co-teacher"},
		{"Michael", "Mike", "O'Brien", "michael.o@schoolofnorthernlights.org", "EA"},
		{"Emily", "Em", "Kim", "emily.k@schoolofnorthernlights.org", "admin"},
		{"David", "Dave", "Singh", "david.s@schoolofnorthernlights.org", "teacher"},
		{"Jessica", "Jess", "Taylor", "jessica.t@schoolofnorthernlights.org", "co-teacher"},
		{"Daniel", "Dan", "Martinez", "daniel.m@schoolofnorthernlights.org", "EA"},
		{"Ashley", "Ash", "Johnson", "ashley.j@schoolofnorthernlights.org", "teacher"},
		{"Christopher", "Chris", "Lee", "chris.l@schoolofnorthernlights.org", "admin"},
		{"Amanda", "Mandy", "Garcia", "amanda.g@schoolofnorthernlights.org", "co-teacher"},
	}

	ids := make([]string, 0, len(data))
	for _, d := range data {
		cmd := ee.CreateEducatorCommand{
			GivenName:  d.GivenName,
			ChosenName: d.ChosenName,
			FamilyName: d.FamilyName,
			Email:      d.Email,
			Role:       d.Role,
		}
		res, err := ee.CreateEducatorCommandHandler(ctx, cmd, saver)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.EventID)
	}
	return ids, nil
}

func seedEducatorPeriods(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	educatorID,
	periodID string,
) error {
	cmd := epe.AddEducatorToPeriodCommand{
		PeriodID:   periodID,
		EducatorID: educatorID,
	}
	_, err := epe.AddEducatorToPeriodCommandHandler(ctx, cmd, saver, retriever)
	return err
}

func seedStudentPeriods(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	studentID,
	periodID string,
) error {
	cmd := spe.AddStudentToPeriodCommand{
		PeriodID:  periodID,
		StudentID: studentID,
	}
	_, err := spe.AddStudentToPeriodCommandHandler(ctx, cmd, saver, retriever)
	return err
}

func parseTimeOnly(s string) sharedmodels.TimeOnly {
	t, _ := time.Parse("15:04", s)
	return sharedmodels.TimeOnly(t)
}
