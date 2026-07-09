package httpserver

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

// sends a non-blocking signal
// if there is already a signal pending, drop remaining signals
type DedupeNotifier struct {
	ch chan struct{}
}

// creates a notifier with a 1-capacity buffer
func NewDedupeNotifier() *DedupeNotifier {
	return &DedupeNotifier{
		ch: make(chan struct{}, 1),
	}
}

// notify sends a signal to the buffer without blocking
func (n *DedupeNotifier) Notify() {
	select {
	case n.ch <- struct{}{}:
	default:
	}
}

// returns the recieve-only channel
// use `<-notifier.Signal()` to be notified of updates
func (n *DedupeNotifier) Signal() <-chan struct{} {
	return n.ch
}

func writeSSE(w http.ResponseWriter, r *http.Request, fn func(*datastar.ServerSentEventGenerator) error) {
	_ = fn(newSSE(w, r))
}

func newSSE(w http.ResponseWriter, r *http.Request) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(w, r, datastar.WithCompression())
}

func emptySSE(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
			return flashError(sse, err.Error())
		})
		return
	}
	writeSSE(w, r, clearFlash)
}

func clearSignals(signals any, sse *datastar.ServerSentEventGenerator) error {
	return sse.MarshalAndPatchSignals(signals)
}

func alert(sse *datastar.ServerSentEventGenerator, message string) error {
	return flashError(sse, message)
}

func flashError(sse *datastar.ServerSentEventGenerator, message string) error {
	return sse.MarshalAndPatchSignals(map[string]string{"flashMessage": message})
}

func clearFlash(sse *datastar.ServerSentEventGenerator) error {
	return sse.MarshalAndPatchSignals(map[string]string{"flashMessage": ""})
}

func patchTempl(w http.ResponseWriter, r *http.Request, component templ.Component, opts ...datastar.PatchElementOption) {
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchElementTempl(component, opts...)
	})
}
