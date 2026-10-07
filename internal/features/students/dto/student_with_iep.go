package dto

import (
	iepModels "seek/internal/features/ieps/models"
	"seek/internal/features/students/models"
)

type StudentWithIEP struct {
	models.Student
	IEP iepModels.IEP `json:"iep"`
}
