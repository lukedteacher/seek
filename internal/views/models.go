package views

import (
	"os"
	"time"
)

type Student struct {
	Id             string			`db:"id"`
	FirstName      string			`db:"first_name"`
	ChosenName     *string			`db:"chosen_name"`
	LastName       string			`db:"last_name"`
	Grade          int64				`db:"grade"`
	Homeroom       string			`db:"homeroom"`
	CaseManager    *string			`db:"case_manager"`
	CreatedAt      time.Time	`db:"created_at"`
	UpdatedAt      time.Time	`db:"updated_at"`
}

type DataView int

const (
	DataViewCard	DataView = iota
	DataViewList
)

func HotReloadSSE() string {
	return "@get('/reload', {retryMaxCount: 1000, retryInterval: 20, retryMaxWaitMs: 200})"
}

func IsDevelopment() bool {
	return os.Getenv("NODE_ENV") != "production"
}