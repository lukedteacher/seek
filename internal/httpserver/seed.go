package httpserver

import (
	"context"
	"net/http"
	"time"

	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	ee "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
)

func (s Server) seedData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := auth.RegisterUserCommandHandler(ctx, auth.RegisterUserCommand{
		Email:    "luke@lukeout.world",
		Password: "fart123",
		Metadata: eventstore.HTTPCommandMetadata(r, ""),
	}, s.EventSaver, s.EventRetriever, s.PIIKeys)
	if err != nil {
		s.Logger.ErrorContext(ctx, "seed data user", "err", err)
		return
	}

	if err := seedStudents(ctx, s); err != nil {
		s.Logger.ErrorContext(ctx, "seed data students", "err", err)
	}

	if err := seedPeriods(ctx, s); err != nil {
		s.Logger.ErrorContext(ctx, "seed data periods", "err", err)
	}

	if err := seedEducators(ctx, s); err != nil {
		s.Logger.ErrorContext(ctx, "seed data educators", "err", err)
	}
}

func seedStudents(ctx context.Context, s Server) error {
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

	for _, datum := range data {
		cmd := se.CreateStudentCommand{
			GivenName:   datum.GivenName,
			ChosenName:  datum.ChosenName,
			FamilyName:  datum.FamilyName,
			Email:       datum.Email,
			Grade:       datum.Grade,
			Homeroom:    datum.Homeroom,
			CaseManager: datum.CaseManager,
		}
		_, err := se.CreateStudentCommandHandler(ctx, cmd, s.EventSaver)
		if err != nil {
			s.Logger.ErrorContext(ctx, "seed student", "err", err)
			return err
		}
	}
	return nil
}

func seedPeriods(ctx context.Context, s Server) error {
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

	for _, datum := range data {
		cmd := pe.CreatePeriodCommand{
			Title:       datum.Title,
			ServiceType: datum.ServiceType,
			StartTime:   datum.StartTime,
			Duration:    datum.Duration,
			DaysBitmask: datum.DaysBitmask,
		}
		_, err := pe.CreatePeriodCommandHandler(ctx, cmd, s.EventSaver)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedEducators(ctx context.Context, s Server) error {
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

	for _, datum := range data {
		cmd := ee.CreateEducatorCommand{
			GivenName:  datum.GivenName,
			ChosenName: datum.ChosenName,
			FamilyName: datum.FamilyName,
			Email:      datum.Email,
			Role:       datum.Role,
		}
		_, err := ee.CreateEducatorCommandHandler(ctx, cmd, s.EventSaver)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTimeOnly(s string) sharedmodels.TimeOnly {
	t, _ := time.Parse("15:04", s)
	return sharedmodels.TimeOnly(t)
}
