package events

import (
	"time"

	"seek/internal/appdb"
	"seek/internal/eventstore"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/ieps/models"
)

type eventType = eventstore.EventType

// event types
const (
	EventIEPAddedToStudent     eventType = "iep_added_to_student_event"
	EventIEPRemovedFromStudent eventType = "iep_removed_from_student_event"
	EventIEPUpdated            eventType = "iep_updated_event"
	EventIEPArchived           eventType = "iep_archived_event"
	EventIEPDeleted            eventType = "iep_deleted_event"
)

// event fields for event IDs
const (
	FieldIEPAddedToStudentEventID     = "iep_added_to_student_event_id"
	FieldIEPRemovedFromStudentEventID = "iep_removed_from_student_event_id"
	FieldIEPUpdatedEventID            = "iep_updated_event_id"
	FieldIEPArchivedEventID           = "iep_archived_event_id"
	FieldIEPDeletedEventID            = "iep_deleted_event_id"
)

// event fields
const (
	FieldIEPID          = "id"
	FieldIEPStudentID   = "student_id"
	FieldIEPStartDate   = "start_date"
	FieldIEPEndDate     = "end_date"
	FieldIEPAmendedDate = "amended_date"
	FieldIEPAddedAt     = "added_at"
	FieldIEPUpdatedAt   = "updated_at"
	FieldIEPArchivedAt  = "archived_at"
	FieldIEPDeletedAt   = "deleted_at"
	FieldIEPScopeID     = "scope.iep_id"
)

type IEPState struct {
	ID                    string `json:"id"`
	StudentID             string `json:"student_id"`
	Disability1           int64  `json:"disability_1,omitempty"`
	Disability2           int64  `json:"disability_2,omitempty"`
	FederalSetting        int64  `json:"federal_setting,omitempty"`
	MeetingDate           string `json:"meeting_date,omitempty"`
	IEPDueDate            string `json:"iep_due_date,omitempty"`
	LastEvalDate          string `json:"last_eval_date,omitempty"`
	EvalDueDate           string `json:"eval_due_date,omitempty"`
	AmendedDate           string `json:"amended_date,omitempty"`
	IEPType               int64  `json:"iep_type,omitempty"`
	SpecialTransportation int64  `json:"special_transportation,omitempty"`
	AddedAt               string `json:"added_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
	ArchivedAt            string `json:"archived_at,omitempty"`
	DeletedAt             string `json:"deleted_at,omitempty"`
}

func NewStateFromModel(m models.IEP) IEPState {
	return IEPState{
		ID:                    m.ID,
		StudentID:             m.StudentID,
		Disability1:           int64(m.Disability1),
		Disability2:           int64(m.Disability2),
		FederalSetting:        int64(m.FederalSetting),
		MeetingDate:           m.MeetingDate.String(),
		IEPDueDate:            m.IEPDueDate.String(),
		LastEvalDate:          m.LastEvalDate.String(),
		EvalDueDate:           m.EvalDueDate.String(),
		AmendedDate:           m.AmendedDate.String(),
		IEPType:               int64(m.IEPType),
		SpecialTransportation: boolToInt64(m.SpecialTransportation),
		AddedAt:               appdb.SQLTime(m.CreatedAt),
		UpdatedAt:             appdb.SQLTime(m.UpdatedAt),
	}
}

type IEPFlat struct {
	ID                    string `json:"iep.id"`
	StudentID             string `json:"iep.student_id"`
	Disability1           int64  `json:"iep.disability_1,omitempty"`
	Disability2           int64  `json:"iep.disability_2,omitempty"`
	FederalSetting        int64  `json:"iep.federal_setting,omitempty"`
	MeetingDate           string `json:"iep.meeting_date,omitempty"`
	IEPDueDate            string `json:"iep.iep_due_date,omitempty"`
	LastEvalDate          string `json:"iep.last_eval_date,omitempty"`
	EvalDueDate           string `json:"iep.eval_due_date,omitempty"`
	AmendedDate           string `json:"iep.amended_date,omitempty"`
	IEPType               int64  `json:"iep.iep_type,omitempty"`
	SpecialTransportation int64  `json:"iep.special_transportation,omitempty"`
	AddedAt               string `json:"iep.added_at,omitempty"`
	UpdatedAt             string `json:"iep.updated_at,omitempty"`
	ArchivedAt            string `json:"iep.archived_at,omitempty"`
	DeletedAt             string `json:"iep.deleted_at,omitempty"`
}

func NewModelFromFlat(f IEPFlat) models.IEP {
	return models.IEP{
		ID:                    f.ID,
		StudentID:             f.StudentID,
		Disability1:           models.DisabilityCode(f.Disability1),
		Disability2:           models.DisabilityCode(f.Disability2),
		FederalSetting:        int(f.FederalSetting),
		MeetingDate:           sharedmodels.DateOnly(parseDBTime(f.MeetingDate)),
		IEPDueDate:            sharedmodels.DateOnly(parseDBTime(f.IEPDueDate)),
		LastEvalDate:          sharedmodels.DateOnly(parseDBTime(f.LastEvalDate)),
		EvalDueDate:           sharedmodels.DateOnly(parseDBTime(f.EvalDueDate)),
		AmendedDate:           sharedmodels.DateOnly(parseDBTime(f.AmendedDate)),
		IEPType:               models.IEPType(f.IEPType),
		SpecialTransportation: int64ToBool(f.SpecialTransportation),
		CreatedAt:             parseDBTime(f.AddedAt),
		UpdatedAt:             parseDBTime(f.UpdatedAt),
	}
}

type StudentState struct {
	created      bool
	archived     bool
	deleted      bool
	hasActiveIEP bool
}

func (student StudentState) isActive() bool {
	if student.created && !student.archived && !student.deleted {
		return true
	}
	return false
}

func (s StudentState) hasIEP() bool {
	if s.hasActiveIEP {
		return true
	}
	return false
}

type IEPAddedToStudentEvent struct {
	ID    string   `json:"iep_added_to_student_event_id"`
	IEP   IEPState `json:"iep"`
	Scope IEPScope `json:"scope"`
}

type IEPUpdatedEvent struct {
	ID    string   `json:"iep_updated_event_id"`
	IEP   IEPState `json:"iep"`
	Scope IEPScope `json:"scope"`
}

type IEPArchivedEvent struct {
	ID         string   `json:"iep_archived_event_id"`
	IEPID      string   `json:"iep.id"`
	ArchivedAt string   `json:"iep.archived_at"`
	Scope      IEPScope `json:"scope"`
}

type IEPDeletedEvent struct {
	ID        string   `json:"iep_deleted_event_id"`
	IEPID     string   `json:"iep.id"`
	DeletedAt string   `json:"iep.deleted_at"`
	Scope     IEPScope `json:"scope"`
}

type IEPScope struct {
	IEPID     string `json:"iep_id"`
	StudentID string `json:"student_id"`
}

func NewIEPAddedToStudentEvent(
	cmd AddIEPToStudentCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	now := time.Now()
	cmd.IEP.CreatedAt = now
	cmd.IEP.UpdatedAt = now
	state := NewStateFromModel(cmd.IEP)
	event := IEPAddedToStudentEvent{
		ID:    cmd.IEP.ID,
		IEP:   state,
		Scope: iepScope(cmd.IEP.ID, cmd.IEP.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   cmd.IEP.ID,
		EventType: EventIEPAddedToStudent,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewIEPUpdatedEvent(
	eventID string,
	cmd UpdateIEPCommand,
	query eventstore.Query,
) eventstore.DomainEvent {
	now := time.Now()
	cmd.IEP.UpdatedAt = now
	state := NewStateFromModel(cmd.IEP)
	event := IEPUpdatedEvent{
		ID:    eventID,
		IEP:   state,
		Scope: iepScope(cmd.IEP.ID, cmd.IEP.StudentID),
	}
	metadata := metadataWithQuery(cmd.Metadata, query)
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewIEPDeletedEvent(
	eventID string,
	IEPID string,
	studentID string,
	deletedAt time.Time,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := IEPDeletedEvent{
		ID:    eventID,
		Scope: iepScope(IEPID, studentID),
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: EventIEPDeleted,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func iepScope(iepID, studentID string) IEPScope {
	return IEPScope{
		IEPID:     iepID,
		StudentID: studentID,
	}
}

func Channel(id string) string {
	return "ieps." + id
}

func ChannelAll() string {
	return "ieps.>"
}
