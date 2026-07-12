package dto

type StudentView struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	ChosenName  string `json:"chosen_name"`
	LastName    string `json:"last_name"`
	Grade       string `json:"grade"`
	Homeroom    string `json:"homeroom"`
	CaseManager string `json:"case_manager"`
}
