package views

import (
	"fmt"
	"os"
)

type Student struct {
	Id             string			`db:"id"`
	FirstName      string			`db:"first_name"`
	ChosenName     *string			`db:"chosen_name"`
	LastName       string			`db:"last_name"`
	Grade          int64				`db:"grade"`
	Homeroom       string			`db:"homeroom"`
	CaseManager    *string			`db:"case_manager"`
}

type DataView int

const (
	DataViewCard	DataView = iota
	DataViewList
)

func LongRunningGetSSE(path string) string {
	return fmt.Sprintf("@get('%s', {requestCancellation: 'disabled', retryMaxCount: 1000, retryInterval: 1000, retryMaxWaitMs: 5000})", path)
}

func HotReloadSSE() string {
	return "@get('/reload', {retryMaxCount: 1000, retryInterval: 20, retryMaxWaitMs: 200})"
}

func IsDevelopment() bool {
	return os.Getenv("NODE_ENV") != "production"
}