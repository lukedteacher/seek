package models

import (
	educatorModels "seek/internal/features/educators/models"
	studentModels "seek/internal/features/students/models"
)

type CaseManager struct {
	educatorModels.Educator
	Caseload []studentModels.Student
}
