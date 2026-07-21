package models

type Student struct {
	ID          string  `json:"id" display:"ID"`
	FirstName   string  `json:"first_name" display:"given"`
	ChosenName  string `json:"chosen_name" display:"chosen"`
	LastName    string  `json:"last_name" display:"family"`
	Grade       int64   `json:"grade" display:"grade" format:"GradeOrdinal" renderer:"badge"`
	Homeroom    string  `json:"homeroom" display:"homeroom"`
	CaseManager string `json:"case_manager" display:"case manager"`
}