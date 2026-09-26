package compositedto

import (
	"seek/internal/features/homerooms/models"
)

type HomeroomWithData struct {
	models.Homeroom
	Educators []EducatorWithData
	Students  []StudentWithData
}
