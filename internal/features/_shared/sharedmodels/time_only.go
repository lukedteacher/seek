package sharedmodels

import (
	"encoding/json"
	"fmt"
	"time"
)

type TimeOnly time.Time

func (t TimeOnly) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t TimeOnly) Add(minutes int) TimeOnly {
	return TimeOnly(time.Time(t).Add(time.Duration(minutes) * time.Minute))
}

func (t TimeOnly) Format(layout string) string {
	return time.Time(t).Format(layout)
}

func (t TimeOnly) Time() time.Time {
	return time.Time(t)
}

func (t TimeOnly) String() string {
	return time.Time(t).Format("15:04")
}

func (t TimeOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String()) // returns "15:04" as quoted string
}

func (t *TimeOnly) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	layouts := []string{"15:04", time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			*t = TimeOnly(parsed)
			return nil
		}
	}
	return fmt.Errorf("cannot parse time %q", s)
}
