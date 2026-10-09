package sharedmodels

type SortDirection string

const (
	SortDirectionAscending  SortDirection = "ASC"
	SortDirectionDescending SortDirection = "DESC"
)

var SortList = []SortDirection{
	SortDirectionAscending,
	SortDirectionDescending,
}
