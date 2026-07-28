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
			Person: sharedmodels.Person{
				GivenName:  "Emma",
				ChosenName: "Em",
				FamilyName: "Johnson",
			},
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Mr. Davis",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Liam",
				ChosenName: "",
				FamilyName: "Garcia",
			},
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "Ms. Wilson",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Olivia",
				ChosenName: "Liv",
				FamilyName: "Martinez",
			},
			Grade:       2,
			Homeroom:    "Mrs. Lee",
			CaseManager: "",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Noah",
				ChosenName: "",
				FamilyName: "Rodriguez",
			},
			Grade:       1,
			Homeroom:    "Mr. Jones",
			CaseManager: "Ms. Taylor",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Ava",
				ChosenName: "Aves",
				FamilyName: "Williams",
			},
			Grade:       4,
			Homeroom:    "Mrs. Clark",
			CaseManager: "Mr. Harris",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Mason",
				ChosenName: "Mace",
				FamilyName: "Brown",
			},
			Grade:       8,
			Homeroom:    "Ms. White",
			CaseManager: "",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Sophia",
				ChosenName: "Soph",
				FamilyName: "Jones",
			},
			Grade:       0,
			Homeroom:    "Mr. Anderson",
			CaseManager: "Ms. Thomas",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Logan",
				ChosenName: "",
				FamilyName: "Miller",
			},
			Grade:       7,
			Homeroom:    "Mrs. Martinez",
			CaseManager: "Mr. Garcia",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Mia",
				ChosenName: "",
				FamilyName: "Davis",
			},
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Ms. Wilson",
		},
		{
			Person: sharedmodels.Person{
				GivenName:  "Ethan",
				ChosenName: "E",
				FamilyName: "Moore",
			},
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "",
		},
	}

	// helper if CaseManager is *string
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
			println("error in seed data")
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
			println("error in seed data")
			return
		}
	}
}

func parseTimeOnly(s string) sharedmodels.TimeOnly {
	t, _ := time.Parse("15:04", s)
	return sharedmodels.TimeOnly(t)
}
