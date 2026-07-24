package dto

import (
	"seek/internal/features/_shared/shareddto"
	"seek/internal/features/_shared/sharedmodels"
	"seek/internal/features/iep_services/models"
)

func NewIEPServiceDiffTableView(diffs []sharedmodels.Diff[models.IEPService]) shareddto.DiffTableView {
	return shareddto.NewDiffTableView(diffs, IEPServiceTableConfig)
}
