package shareddto

import (
	"seek/internal/features/_shared/sharedmodels"
)

type SelectViewServiceType struct {
	Options []SelectOptionServiceType `json:"options"`
}

type SelectOptionServiceType struct {
	sharedmodels.ServiceType
	Selected bool
}

func NewSelectViewServiceType(
	selected sharedmodels.ServiceType,
) SelectViewServiceType {
	options := make([]SelectOptionServiceType, len(sharedmodels.ServiceTypeList))
	for i, serviceType := range sharedmodels.ServiceTypeList {
		options[i] = NewSelectOptionServiceType(
			serviceType,
			selected == serviceType,
		)
	}
	return SelectViewServiceType{
		Options: options,
	}
}

func NewSelectOptionServiceType(
	st sharedmodels.ServiceType,
	selected bool,
) SelectOptionServiceType {
	return SelectOptionServiceType{
		ServiceType: st,
		Selected:    selected,
	}
}
