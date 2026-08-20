package httpserver

import (
	"net/http"
	"seek/internal/ui/core/coreblocks/toasts"
	"sync"
	"time"

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

type MessageNotifier struct {
	mu      sync.Mutex
	ch      chan []byte
	timer   *time.Timer
	pending []byte // last message to send after coalescing
}

func NewMessageNotifier() *MessageNotifier {
	return &MessageNotifier{
		ch: make(chan []byte, 1), // buffered to avoid blocking
	}
}

func (n *MessageNotifier) Notify(data []byte) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.pending = data
	if n.timer == nil {
		n.timer = time.AfterFunc(10*time.Millisecond, func() {
			n.mu.Lock()
			defer n.mu.Unlock()
			select {
			case n.ch <- n.pending:
			default:
			}
			n.timer = nil
		})
	}
}

func (n *MessageNotifier) Signal() <-chan []byte {
	return n.ch
}

func writeSSE(w http.ResponseWriter, r *http.Request, fn func(*datastar.ServerSentEventGenerator) error) {
	_ = fn(newSSE(w, r))
}

func newSSE(w http.ResponseWriter, r *http.Request) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(w, r, datastar.WithCompression())
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

func toastError(sse *datastar.ServerSentEventGenerator, message string) error {
	return sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantError, message))
}

func toastInfo(sse *datastar.ServerSentEventGenerator, message string) error {
	return sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantInfo, message))
}

func toastSuccess(sse *datastar.ServerSentEventGenerator, message string) error {
	return sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantSuccess, message))
}

func toastWarning(sse *datastar.ServerSentEventGenerator, message string) error {
	return sse.PatchElementTempl(toasts.ToastContainer(toasts.VariantWarning, message))
}

func patchTempl(w http.ResponseWriter, r *http.Request, component templ.Component, opts ...datastar.PatchElementOption) {
	writeSSE(w, r, func(sse *datastar.ServerSentEventGenerator) error {
		return sse.PatchElementTempl(component, opts...)
	})
}
