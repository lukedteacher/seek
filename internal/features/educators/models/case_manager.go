package models

import (
	"seek/internal/features/students/models"
)

type CaseManager struct {
	Educator
	Caseload []models.Student
}
