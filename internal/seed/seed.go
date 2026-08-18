package seed

import (
	"context"
	"strings"
	"time"

	"seek/internal/auth"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/educators/events"
	ee "seek/internal/features/educators/events"
	pe "seek/internal/features/periods/events"
	se "seek/internal/features/students/events"
)

func SeedStudents(
	ctx context.Context,
	saver eventstore.Saver,
) (
	[]string,
	error,
) {
	data := []struct {
		MARSSID    string
		GivenName  string
		ChosenName string
		FamilyName string
		Email      string
		Grade      int
	}{
		{"0625000790560", "Emma", "Em", "Johnson", "emma.j@schoolofnorthernlights.org", 5},
		{"0625000795705", "Liam", "", "Garcia", "liam.g@schoolofnorthernlights.org", 6},
		{"1", "Olivia", "Liv", "Martinez", "olivia.m@schoolofnorthernlights.org", 2},
		{"1", "Noah", "", "Rodriguez", "noah.r@schoolofnorthernlights.org", 1},
		{"1", "Ava", "Aves", "Williams", "ava.w@schoolofnorthernlights.org", 4},
		{"1", "Mason", "Mace", "Brown", "mason.b@schoolofnorthernlights.org", 8},
		{"1", "Sophia", "Soph", "Jones", "sophia.j@schoolofnorthernlights.org", 0},
		{"1", "Logan", "", "Miller", "logan.m@schoolofnorthernlights.org", 7},
		{"1", "Mia", "", "Davis", "mia.d@schoolofnorthernlights.org", 5},
		{"1", "Ethan", "E", "Moore", "ethan.m@schoolofnorthernlights.org", 6},
	}

	ids := make([]string, 0, len(data))
	for _, d := range data {
		cmd := se.CreateStudentCommand{
			MARSSID:    d.MARSSID,
			GivenName:  d.GivenName,
			ChosenName: d.ChosenName,
			FamilyName: d.FamilyName,
			Email:      d.Email,
			Grade:      d.Grade,
		}
		res, err := se.CreateStudentCommandHandler(ctx, cmd, saver)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.EventID)
	}
	return ids, nil
}

func SeedPeriods(
	ctx context.Context,
	saver eventstore.Saver,
) (
	[]string,
	error,
) {
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

func SeedEducators(
	ctx context.Context,
	saver eventstore.Saver,
) (
	[]string,
	error,
) {
	data := []struct {
		GivenName  string
		ChosenName string
		FamilyName string
		Email      string
		Role       string
	}{
		{"luke", "!uke", "earley", "luke.e@schoolofnorthernlights.org", "service provider,resource room teacher"},
		{"Sarah", "Sally", "Chen", "sarah.c@schoolofnorthernlights.org", "co-teacher,case manager"},
		{"Michael", "Mike", "O'Brien", "michael.o@schoolofnorthernlights.org", "educational assistant"},
		{"Emily", "Em", "Kim", "emily.k@schoolofnorthernlights.org", "admin"},
		{"David", "Dave", "Singh", "david.s@schoolofnorthernlights.org", "general education teacher"},
		{"Jessica", "Jess", "Taylor", "jessica.t@schoolofnorthernlights.org", "service provider,co-teacher,case manager"},
		{"Daniel", "Dan", "Martinez", "daniel.m@schoolofnorthernlights.org", "educational assistant"},
		{"Ashley", "Ash", "Johnson", "ashley.j@schoolofnorthernlights.org", "co-teacher,case manager"},
		{"Christopher", "Chris", "Lee", "chris.l@schoolofnorthernlights.org", "admin"},
		{"Amanda", "Mandy", "Garcia", "amanda.g@schoolofnorthernlights.org", "service provider,co-teacher,case manager"},
	}

	ids := make([]string, 0, len(data))
	for _, d := range data {
		cmd := ee.CreateEducatorCommand{
			GivenName:  d.GivenName,
			ChosenName: d.ChosenName,
			FamilyName: d.FamilyName,
			Email:      d.Email,
			Roles:      strings.Split(d.Role, ","),
		}
		res, err := ee.CreateEducatorCommandHandler(ctx, cmd, saver)
		if err != nil {
			return nil, err
		}
		ids = append(ids, res.EventID)
	}
	return ids, nil
}

func parseTimeOnly(s string) sharedmodels.TimeOnly {
	t, _ := time.Parse("15:04", s)
	return sharedmodels.TimeOnly(t)
}

func SeedUsers(
	ctx context.Context,
	saver eventstore.Saver,
	retriever eventstore.Retriever,
	piiKeys auth.SubjectPiiKeyPort,
	educatorReadModel events.ReadModel,
) error {
	educators, err := educatorReadModel.List(ctx)
	if err != nil {
		return err
	}
	for _, educator := range educators {
		_, err := auth.RegisterUserCommandHandler(ctx, auth.RegisterUserCommand{
			Email:    educator.Email,
			Password: "sharingiscaring2026",
		}, saver, retriever, piiKeys)
		if err != nil {
			return err
		}

	}
	return nil
}
