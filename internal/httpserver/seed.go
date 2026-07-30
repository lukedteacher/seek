package httpserver

import (
	"net/http"
	"time"

	"seek/internal/features/_shared/sharedmodels"
	pe "seek/internal/features/periods/events"
	pm "seek/internal/features/periods/models"
	se "seek/internal/features/students/events"
	sm "seek/internal/features/students/models"
)

func (s Server) seedData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	students := []sm.Student{
		{
			MARSSID: "1",
			Person: sharedmodels.Person{
				GivenName:  "Emma",
				ChosenName: "Em",
				FamilyName: "Johnson",
				Email:      "emma.j@schoolofnorthernlights.org",
			},
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Mr. Davis",
		},
		{
			MARSSID: "2",
			Person: sharedmodels.Person{
				GivenName:  "Liam",
				ChosenName: "",
				FamilyName: "Garcia",
				Email:      "liam.g@schoolofnorthernlights.org",
			},
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "Ms. Wilson",
		},
		{
			MARSSID: "3",
			Person: sharedmodels.Person{
				GivenName:  "Olivia",
				ChosenName: "Liv",
				FamilyName: "Martinez",
				Email:      "olivia.m@schoolofnorthernlights.org",
			},
			Grade:       2,
			Homeroom:    "Mrs. Lee",
			CaseManager: "",
		},
		{
			MARSSID: "4",
			Person: sharedmodels.Person{
				GivenName:  "Noah",
				ChosenName: "",
				FamilyName: "Rodriguez",
				Email:      "noah.r@schoolofnorthernlights.org",
			},
			Grade:       1,
			Homeroom:    "Mr. Jones",
			CaseManager: "Ms. Taylor",
		},
		{
			MARSSID: "5",
			Person: sharedmodels.Person{
				GivenName:  "Ava",
				ChosenName: "Aves",
				FamilyName: "Williams",
				Email:      "ava.w@schoolofnorthernlights.org",
			},
			Grade:       4,
			Homeroom:    "Mrs. Clark",
			CaseManager: "Mr. Harris",
		},
		{
			MARSSID: "6",
			Person: sharedmodels.Person{
				GivenName:  "Mason",
				ChosenName: "Mace",
				FamilyName: "Brown",
				Email:      "mason.b@schoolofnorthernlights.org",
			},
			Grade:       8,
			Homeroom:    "Ms. White",
			CaseManager: "",
		},
		{
			MARSSID: "7",
			Person: sharedmodels.Person{
				GivenName:  "Sophia",
				ChosenName: "Soph",
				FamilyName: "Jones",
				Email:      "sophia.j@schoolofnorthernlights.org",
			},
			Grade:       0,
			Homeroom:    "Mr. Anderson",
			CaseManager: "Ms. Thomas",
		},
		{
			MARSSID: "8",
			Person: sharedmodels.Person{
				GivenName:  "Logan",
				ChosenName: "",
				FamilyName: "Miller",
				Email:      "logan.m@schoolofnorthernlights.org",
			},
			Grade:       7,
			Homeroom:    "Mrs. Martinez",
			CaseManager: "Mr. Garcia",
		},
		{
			MARSSID: "9",
			Person: sharedmodels.Person{
				GivenName:  "Mia",
				ChosenName: "",
				FamilyName: "Davis",
				Email:      "mia.d@schoolofnorthernlights.org",
			},
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Ms. Wilson",
		},
		{
			MARSSID: "10",
			Person: sharedmodels.Person{
				GivenName:  "Ethan",
				ChosenName: "E",
				FamilyName: "Moore",
				Email:      "ethan.m@schoolofnorthernlights.org",
			},
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "",
		},
	}

	for _, student := range students {
		command := se.CreateStudentCommand{
			GivenName:   student.GivenName,
			ChosenName:  student.ChosenName,
			FamilyName:  student.FamilyName,
			Grade:       int(student.Grade),
			Homeroom:    student.Homeroom,
			CaseManager: student.CaseManager,
		}
		_, err := se.CreateStudentCommandHandler(
			ctx,
			command,
			s.EventSaver,
		)
		if err != nil {
			s.Logger.Error("seed data student", "err", err)
			return
		}
	}

	periods := []pm.Period{
		{Title: "SEL", ServiceType: "SEL", StartTime: parseTimeOnly("08:00"), Duration: 20, DaysBitmask: 1},
		{Title: "reading", ServiceType: "reading", StartTime: parseTimeOnly("08:55"), Duration: 30, DaysBitmask: 5},
		{Title: "writing", ServiceType: "writing", StartTime: parseTimeOnly("09:50"), Duration: 30, DaysBitmask: 15},
		{Title: "SEL", ServiceType: "SEL", StartTime: parseTimeOnly("10:50"), Duration: 20, DaysBitmask: 8},
		{Title: "executive functioning", ServiceType: "EF", StartTime: parseTimeOnly("11:45"), Duration: 15, DaysBitmask: 19},
		{Title: "OT", ServiceType: "OT", StartTime: parseTimeOnly("13:00"), Duration: 30, DaysBitmask: 18},
		{Title: "speech", ServiceType: "speech", StartTime: parseTimeOnly("13:40"), Duration: 20, DaysBitmask: 22},
		{Title: "speech", ServiceType: "speech", StartTime: parseTimeOnly("14:35"), Duration: 20, DaysBitmask: 29},
		{Title: "math (push-in)", ServiceType: "math", StartTime: parseTimeOnly("13:30"), Duration: 60, DaysBitmask: 10},
		{Title: "reading", ServiceType: "reading", StartTime: parseTimeOnly("14:20"), Duration: 30, DaysBitmask: 17},
	}

	for _, period := range periods {
		command := pe.CreatePeriodCommand{
			Title:       period.Title,
			ServiceType: period.ServiceType,
			StartTime:   period.StartTime,
			Duration:    period.Duration,
			DaysBitmask: period.DaysBitmask,
		}
		_, err := pe.CreatePeriodCommandHandler(
			ctx,
			command,
			s.EventSaver,
		)
		if err != nil {
			s.Logger.Error("seed data period", "err", err)
			return
		}
	}
}

func parseTimeOnly(s string) sharedmodels.TimeOnly {
	t, _ := time.Parse("15:04", s)
	return sharedmodels.TimeOnly(t)
}
