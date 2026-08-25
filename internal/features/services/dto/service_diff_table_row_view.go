package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/services/models"
)

func NewServiceDiffTableView(diffs []sharedmodels.Diff[models.Service]) shareddto.DiffTableView {
	return shareddto.NewDiffTableView(diffs, ServiceTableConfig)
}
