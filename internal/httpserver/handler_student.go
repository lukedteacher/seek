package httpserver

import (
	"net/http"
	"strconv"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/student"
	"seek/internal/views/pages"
	"seek/internal/viewstore"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s Server) studentRoutes(r chi.Router) {
	r.Get("/students", s.students)
	r.Get("/students/stream", s.studentsStream)
	r.Get("/students/{id}", s.studentInfo)
	r.Get("/students/create", s.createStudentForm)
	r.Post("/students/create/validate", s.validateCreateStudent)
	r.Post("/students/create", s.createStudent)
	r.Get("/students/{id}/edit", s.editStudentForm)
	r.Post("/students/{id}/edit/validate", s.validateEditStudent)
	r.Post("/students/{id}/edit", s.editStudent)
	r.Delete("/students/{id}", s.deleteStudent)
}

func (s Server) studentInfo(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "id")
	student, err := s.Students.Get(r.Context(), studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.Student(student).Render(r.Context(), w)
}

func (s Server) students(w http.ResponseWriter, r *http.Request) {
	type Signals struct {
		View int64 `json:"view"`
	}
	signals := &Signals{}
	students, err := s.Students.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	datastar.ReadSignals(r, signals)

	_ = pages.Students(signals.View, students).Render(r.Context(), w)
}

func (s Server) studentsStream(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	ctx := r.Context()

	watcher, err := s.ViewStore.Watch(ctx, "students", viewstore.WatchOptions{IgnoreDeletes: true})
	if err != nil {
		_ = alert(sse, err.Error())
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watcher.Updates():
			if !ok {
				return
			}
			students, err := s.Students.List(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			page := pages.Students(0, students)
			if err := sse.PatchElementTempl(page); err != nil {
				return
			}
		}
	}
}

// GET request to /students/create
func (s Server) createStudentForm(w http.ResponseWriter, r *http.Request) {
	emptyStudent := models.Student{}
	validation := student.Validate(nil)
	_ = pages.CreateStudent(emptyStudent, validation, "").Render(r.Context(), w)
}

// POST request to /students/create/validate
func (s Server) validateCreateStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		FirstName   string `json:"first_name"`
		ChosenName  string `json:"chosen_name"`
		LastName    string `json:"last_name"`
		Grade       string `json:"grade"`
		Homeroom    string `json:"homeroom"`
		CaseManager string `json:"case_manager"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	selectedGrade := signals.Grade

	var grade int64
	if signals.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Grade, 10, 64)
	}

	studentToValidate := models.Student{
		FirstName:   signals.FirstName,
		ChosenName:  &signals.ChosenName,
		LastName:    signals.LastName,
		Grade:       grade,
		Homeroom:    signals.Homeroom,
		CaseManager: &signals.CaseManager,
	}
	
	validation := student.Validate(&studentToValidate)
	_ = pages.CreateStudent(studentToValidate, validation, selectedGrade).Render(context, w)
}

// POST request to /student/create
func (s Server) createStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		FirstName   string `json:"first_name"`
		ChosenName  string `json:"chosen_name"`
		LastName    string `json:"last_name"`
		Grade       string `json:"grade"`
		Homeroom    string `json:"homeroom"`
		CaseManager string `json:"case_manager"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	var grade int64
	if signals.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Grade, 10, 64)
	}

	_, err := student.CreateStudentCommandHandler(context, student.CreateStudentCommand{
		FirstName:   signals.FirstName,
		ChosenName:  signals.ChosenName,
		LastName:    signals.LastName,
		Grade:       grade,
		Homeroom:    signals.Homeroom,
		CaseManager: signals.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}

	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return clearSignals(&Signals{}, sse)
	})
}

// GET request to /students/{id}
func (s Server) editStudentForm(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		FirstName   string `json:"first_name"`
		ChosenName  string `json:"chosen_name"`
		LastName    string `json:"last_name"`
		Grade       string  `json:"grade"`
		Homeroom    string `json:"homeroom"`
		CaseManager string `json:"case_manager"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	studentID := chi.URLParam(r, "id")
	studentRes, err := s.Students.Get(context, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	var selectedGrade string = "-1"
	// if the user has selected a teacher, use that as the default selection
	// otherwise use the schedule's teacher ID data
	if signals.Grade != "" && studentRes.GradeString() != signals.Grade {
		selectedGrade = signals.Grade
	} else {
		selectedGrade = studentRes.GradeString()
	}

	var grade int64
	if signals.Grade == "" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Grade, 10, 64)
	}

	model := models.Student{
		Id:          studentID,
		FirstName:   signals.FirstName,
		ChosenName:  &signals.ChosenName,
		LastName:    signals.LastName,
		Grade:       grade,
		Homeroom:    signals.Homeroom,
		CaseManager: &signals.CaseManager,
	}

	validation := student.Validate(&model)
	_ = pages.EditStudent(*studentRes, validation, selectedGrade).Render(context, w)
}

// POST request to /students/{id}/edit/validate
func (s Server) validateEditStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		FirstName   string `json:"first_name"`
		ChosenName  string `json:"chosen_name"`
		LastName    string `json:"last_name"`
		Grade       string `json:"grade"`
		Homeroom    string `json:"homeroom"`
		CaseManager string `json:"case_manager"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	studentID := chi.URLParam(r, "id")
	studentRes, err := s.Students.Get(context, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		println(err.Error())
		return
	}

	var selectedGrade string = "-1"
	// if the user has selected a teacher, use that as the default selection
	// otherwise use the schedule's teacher ID data
	if signals.Grade != "" && studentRes.GradeString() != signals.Grade {
		selectedGrade = signals.Grade
	} else {
		selectedGrade = studentRes.GradeString()
	}

	var grade int64
	if signals.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Grade, 10, 64)
	}

	model := models.Student{
		Id:          studentID,
		FirstName:   signals.FirstName,
		ChosenName:  &signals.ChosenName,
		LastName:    signals.LastName,
		Grade:       grade,
		Homeroom:    signals.Homeroom,
		CaseManager: &signals.CaseManager,
	}
	
	validation := student.Validate(&model)
	_ = pages.EditStudent(*studentRes, validation, selectedGrade).Render(context, w)
}

// POST request to /students/{id}/edit
func (s Server) editStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	type Signals struct {
		FirstName   string `json:"first_name"`
		ChosenName  string `json:"chosen_name"`
		LastName    string `json:"last_name"`
		Grade       int64  `json:"grade"`
		Homeroom    string `json:"homeroom"`
		CaseManager string `json:"case_manager"`
	}
	signals := &Signals{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	studentID := chi.URLParam(r, "id")
	result, err := student.UpdateStudentCommandHandler(context, student.UpdateStudentCommand{
		Id:          studentID,
		FirstName:   signals.FirstName,
		ChosenName:  signals.ChosenName,
		LastName:    signals.LastName,
		Grade:       signals.Grade,
		Homeroom:    signals.Homeroom,
		CaseManager: signals.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Skipped == true {
		println("update skipped")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST request to /students/{id}/delete
func (s Server) deleteStudent(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "id")
	_, err := student.DeleteStudentCommandHandler(r.Context(), student.DeleteStudentCommand{
		StudentID: studentID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
