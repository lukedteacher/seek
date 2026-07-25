package dto

type StudentWithScheduleView struct {
	ID          string       `json:"id"`
	GivenName   string       `json:"first_name"`
	ChosenName  string       `json:"chosen_name"`
	FamilyName  string       `json:"last_name"`
	Grade       string       `json:"grade"`
	Homeroom    string       `json:"homeroom"`
	CaseManager string       `json:"case_manager"`
	Schedule    ScheduleView `json:"schedule"`
}
