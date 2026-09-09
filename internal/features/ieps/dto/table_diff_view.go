package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/ieps/models"
)

func NewServiceDiffTableView(diffs []sharedmodels.Diff[models.IEP]) shareddto.DiffTableView {
	return shareddto.NewDiffTableView(diffs, IEPTableConfig)
}
