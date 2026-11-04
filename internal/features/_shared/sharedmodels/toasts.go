package sharedmodels

import "time"

const (
	ToastLifetime = 5 * time.Second
	ToastMax      = 5
)

type Toast struct {
	ID       string        `json:"id"`
	Kind     string        `json:"kind"`
	Message  string        `json:"message"`
	SentAt   time.Time     `json:"sent_at"`
	Duration time.Duration `json:"duration"`
}

func (t Toast) Live(now time.Time) bool {
	return now.Before(t.SentAt.Add(t.Duration))
}
