package httpserver

import (
	"net/http"
	"strconv"

	"seek/internal/domain/models"
	"seek/internal/eventstore"
	"seek/internal/features/students/events"
	"seek/internal/features/students/pages"
	"seek/internal/views/dto"
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

	_ = pages.View(student).Render(r.Context(), w)
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

	_ = pages.List(signals.View, students).Render(r.Context(), w)
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
			page := pages.List(0, students)
			if err := sse.PatchElementTempl(page); err != nil {
				return
			}
		}
	}
}

// GET request to /students/create
func (s Server) createStudentForm(w http.ResponseWriter, r *http.Request) {
	emptyStudent := models.Student{}
	validation := events.Validate(nil)
	_ = pages.Create(emptyStudent, validation, "").Render(r.Context(), w)
}

// POST request to /students/create/validate
func (s Server) validateCreateStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println(err.Error())
		return
	}

	selectedGrade := signals.Student.Grade

	var grade int64
	if signals.Student.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Student.Grade, 10, 64)
	}

	studentToValidate := models.Student{
		FirstName:   signals.Student.FirstName,
		ChosenName:  &signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
		Grade:       grade,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: &signals.Student.CaseManager,
	}

	validation := events.Validate(&studentToValidate)
	_ = pages.Create(studentToValidate, validation, selectedGrade).Render(ctx, w)
}

// POST request to /student/create
func (s Server) createStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		println(err.Error())
		return
	}

	var grade int64
	if signals.Student.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Student.Grade, 10, 64)
	}

	_, err := events.CreateStudentCommandHandler(ctx, events.CreateStudentCommand{
		FirstName:   signals.Student.FirstName,
		ChosenName:  signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
		Grade:       grade,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver)
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
}

// GET request to /students/{id}/edit
func (s Server) editStudentForm(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
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
	// if the user has selected a grade, use that as the default selection
	// otherwise use the students existing grade data
	if signals.Student.Grade != "" && studentRes.GradeString() != signals.Student.Grade {
		selectedGrade = signals.Student.Grade
	} else {
		selectedGrade = studentRes.GradeString()
	}

	var grade int64
	if signals.Student.Grade == "" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Student.Grade, 10, 64)
	}

	model := models.Student{
		Id:          studentID,
		FirstName:   signals.Student.FirstName,
		ChosenName:  &signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
		Grade:       grade,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: &signals.Student.CaseManager,
	}

	validation := events.Validate(&model)
	_ = pages.Edit(*studentRes, validation, selectedGrade).Render(context, w)
}

// POST request to /students/{id}/edit/validate
func (s Server) validateEditStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
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
	// if the user has selected a grade, use that as the default selection
	// otherwise use the students existing grade data
	if signals.Student.Grade != "" && studentRes.GradeString() != signals.Student.Grade {
		selectedGrade = signals.Student.Grade
	} else {
		selectedGrade = studentRes.GradeString()
	}

	var grade int64
	if signals.Student.Grade == "select a grade" {
		grade = -1
	} else {
		grade, _ = strconv.ParseInt(signals.Student.Grade, 10, 64)
	}

	model := models.Student{
		Id:          studentID,
		FirstName:   signals.Student.FirstName,
		ChosenName:  &signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
		Grade:       grade,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: &signals.Student.CaseManager,
	}

	validation := events.Validate(&model)
	_ = pages.Edit(*studentRes, validation, selectedGrade).Render(context, w)
}

// POST request to /students/{id}/edit
func (s Server) editStudent(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	signals := &struct {
		Student dto.StudentView `json:"student"`
	}{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gradeInt64, err := strconv.ParseInt(signals.Student.Grade, 10, 64)
	if err != nil {
		println("convert int64 error: ", err.Error())
	}

	studentID := chi.URLParam(r, "id")
	result, err := events.UpdateStudentCommandHandler(context, events.UpdateStudentCommand{
		Id:          studentID,
		FirstName:   signals.Student.FirstName,
		ChosenName:  signals.Student.ChosenName,
		LastName:    signals.Student.LastName,
		Grade:       gradeInt64,
		Homeroom:    signals.Student.Homeroom,
		CaseManager: signals.Student.CaseManager,
		Metadata:    eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result.Skipped == true {
		println("update skipped")
		return
	}
}

// POST request to /students/{id}/delete
func (s Server) deleteStudent(w http.ResponseWriter, r *http.Request) {
	studentID := chi.URLParam(r, "id")
	_, err := events.DeleteStudentCommandHandler(r.Context(), events.DeleteStudentCommand{
		StudentID: studentID,
		Metadata:  eventstore.HTTPCommandMetadata(r),
	}, s.EventSaver, s.EventRetriever)
	emptySSE(w, r, err)
}
