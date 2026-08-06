package sharedmodels

type Location struct {
	Name       string
	Short      string
	RoomNumber string
}

var LocationList = []Location{
	Location{
		Name:       "resource room",
		Short:      "res",
		RoomNumber: "1",
	},
	Location{
		Name:       "gym",
		Short:      "gym",
		RoomNumber: "2",
	},
	Location{
		Name:       "conference room",
		Short:      "conf",
		RoomNumber: "3",
	},
	Location{
		Name:       "scott's classroom",
		Short:      "scott",
		RoomNumber: "4",
	},
}
