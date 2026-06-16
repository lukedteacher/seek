package views

import (
	"os"
)

type DataView int

const (
	DataViewCard	DataView = iota
	DataViewList
)

func HotReloadSSE() string {
	println("hot reload called")
	return "@get('/reload', {retryMaxCount: 1000, retryInterval: 20, retryMaxWaitMs: 200})"
}

func IsDevelopment() bool {
	return os.Getenv("NODE_ENV") != "production"
}