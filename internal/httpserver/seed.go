package httpserver

import (
	"net/http"

	pe "seek/internal/features/periods/events"
	pm "seek/internal/features/periods/models"
	se "seek/internal/features/students/events"
	sm "seek/internal/features/students/models"
)

func (s Server) seedData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	students := []sm.Student{
		{
			FirstName:   "Emma",
			ChosenName:  "Em",
			LastName:    "Johnson",
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Mr. Davis",
		},
		{
			FirstName:   "Liam",
			ChosenName:  "",
			LastName:    "Garcia",
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "Ms. Wilson",
		},
		{
			FirstName:   "Olivia",
			ChosenName:  "Liv",
			LastName:    "Martinez",
			Grade:       2,
			Homeroom:    "Mrs. Lee",
			CaseManager: "",
		},
		{
			FirstName:   "Noah",
			ChosenName:  "",
			LastName:    "Rodriguez",
			Grade:       1,
			Homeroom:    "Mr. Jones",
			CaseManager: "Ms. Taylor",
		},
		{
			FirstName:   "Ava",
			ChosenName:  "Aves",
			LastName:    "Williams",
			Grade:       4,
			Homeroom:    "Mrs. Clark",
			CaseManager: "Mr. Harris",
		},
		{
			FirstName:   "Mason",
			ChosenName:  "Mace",
			LastName:    "Brown",
			Grade:       8,
			Homeroom:    "Ms. White",
			CaseManager: "",
		},
		{
			FirstName:   "Sophia",
			ChosenName:  "Soph",
			LastName:    "Jones",
			Grade:       0,
			Homeroom:    "Mr. Anderson",
			CaseManager: "Ms. Thomas",
		},
		{
			FirstName:   "Logan",
			ChosenName:  "",
			LastName:    "Miller",
			Grade:       7,
			Homeroom:    "Mrs. Martinez",
			CaseManager: "Mr. Garcia",
		},
		{
			FirstName:   "Mia",
			ChosenName:  "",
			LastName:    "Davis",
			Grade:       5,
			Homeroom:    "Ms. Smith",
			CaseManager: "Ms. Wilson",
		},
		{
			FirstName:   "Ethan",
			ChosenName:  "E",
			LastName:    "Moore",
			Grade:       6,
			Homeroom:    "Mr. Brown",
			CaseManager: "",
		},
	}

	for _, student := range students {
		command := se.CreateStudentCommand{
			FirstName:   student.FirstName,
			ChosenName:  student.ChosenName,
			LastName:    student.LastName,
			Grade:       student.Grade,
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
		{Title: "SEL", StartTime: "08:00", Duration: 20, Days: 1},
		{Title: "reading", StartTime: "08:55", Duration: 30, Days: 5},
		{Title: "writing", StartTime: "09:50", Duration: 30, Days: 15},
		{Title: "SEL", StartTime: "10:50", Duration: 20, Days: 8},
		{Title: "ex fun", StartTime: "11:45", Duration: 15, Days: 19},
		{Title: "OT", StartTime: "13:00", Duration: 30, Days: 18},
		{Title: "speech", StartTime: "13:40", Duration: 20, Days: 22},
		{Title: "speech", StartTime: "14:35", Duration: 20, Days: 29},
		{Title: "math (push-in)", StartTime: "13:30", Duration: 60, Days: 10},
		{Title: "reading", StartTime: "14:20", Duration: 30, Days: 17},
	}

	for _, period := range periods {
		command := pe.CreatePeriodCommand{
			Title:     period.Title,
			StartTime: period.StartTime,
			Duration:  period.Duration,
			Days:      period.Days,
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
