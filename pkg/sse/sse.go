package sse

import (
	"fmt"
	"os"
)

func PostFormSSE(path string) string {
	return fmt.Sprintf("@post('%s', {contentType: 'form'})", path)
}

func LongRunningGetSSE(path string) string {
	return fmt.Sprintf("@get('%s', {requestCancellation: 'disabled', retryMaxCount: 1000, retryInterval: 1000, retryMaxWaitMs: 5000})", path)
}

func HotReloadSSE() string {
	return "@get('/reload', {retryMaxCount: 1000, retryInterval: 20, retryMaxWaitMs: 200})"
}

func IsDevelopment() bool {
	return os.Getenv("NODE_ENV") != "production"
}
