package dto

import (
	educatorModels "seek/internal/features/educators/models"
	periodModels "seek/internal/features/periods/models"
	studentModels "seek/internal/features/students/models"
)

type EducatorWithPeriods struct {
	educatorModels.Educator
	Periods           []periodModels.Period
	PeriodStudentsMap map[string][]studentModels.Student
}
