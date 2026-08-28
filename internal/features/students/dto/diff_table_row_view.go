package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/students/models"
)

func NewStudentsDiffTableView(diffs []sharedmodels.Diff[models.Student]) shareddto.DiffTableView {
	return shareddto.NewDiffTableView(diffs, StudentTableConfig)
}
