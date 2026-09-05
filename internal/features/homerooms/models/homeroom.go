package models

type Homeroom struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	LocationID  string   `json:"locationID"`
	EducatorIDs []string `json:"educator_ids"`
	StudentIDs  []string `json:"student_ids"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

func NewHomeroom() Homeroom {
	return Homeroom{}
}
