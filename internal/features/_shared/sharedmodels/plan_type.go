package sharedmodels

type PlanType int

const (
	PlanTypeNone PlanType = iota
	PlanTypeChildFind
	PlanType504
	PlanTypeIEP
)

var PlanTypeList = []PlanType{
	PlanTypeNone,
	PlanTypeChildFind,
	PlanType504,
	PlanTypeIEP,
}

func (pt PlanType) String() string {
	stringMap := map[PlanType]string{
		0: "none",
		1: "child find",
		2: "504",
		3: "IEP",
	}
	return stringMap[pt]
}
