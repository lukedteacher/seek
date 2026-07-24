package sharedmodels

type Diff[T any] struct {
	Key           string
	Status        DiffStatus
	Old           *T
	New           *T
	ChangedFields []string
}

type DiffStatus int

const (
	DiffSame DiffStatus = iota
	DiffUpdated
	DiffAdded
	DiffRemoved
)

var DiffStatusStringMap = map[DiffStatus]string{
	DiffSame:    "same",
	DiffUpdated: "updated",
	DiffAdded:   "added",
	DiffRemoved: "removed",
}

func (ds DiffStatus) String() string {
	return DiffStatusStringMap[ds]
}

// compare two slices of any type, returning diffs.
// keyFn generates a unique key for each item.
// diffFields compares two items and returns a list of changed field names.
func Compare[T any](
	db []*T,
	csv []*T,
	keyFn func(*T) string,
	diffFields func(*T, *T) []string,
) []Diff[T] {
	dbMap := make(map[string]*T)
	for _, s := range db {
		dbMap[keyFn(s)] = s
	}

	csvKeys := make(map[string]bool)
	var diffs []Diff[T]

	for _, csvItem := range csv {
		key := keyFn(csvItem)
		csvKeys[key] = true
		if dbItem, exists := dbMap[key]; exists {
			if fields := diffFields(dbItem, csvItem); len(fields) > 0 {
				diffs = append(diffs, Diff[T]{
					Key:           key,
					Status:        DiffUpdated,
					Old:           dbItem,
					New:           csvItem,
					ChangedFields: fields,
				})
			} else {
				diffs = append(diffs, Diff[T]{
					Key:    key,
					Status: DiffSame,
					New:    csvItem,
				})
			}
		} else {
			diffs = append(diffs, Diff[T]{
				Key:    key,
				Status: DiffAdded,
				New:    csvItem,
			})
		}
	}

	for key, dbItem := range dbMap {
		if !csvKeys[key] {
			diffs = append(diffs, Diff[T]{
				Key:    key,
				Status: DiffRemoved,
				Old:    dbItem,
			})
		}
	}
	return diffs
}
