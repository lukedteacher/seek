package sharedmodels

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type GradesBitmask int

const (
	BitGradeK = 1 << iota
	BitGrade1
	BitGrade2
	BitGrade3
	BitGrade4
	BitGrade5
	BitGrade6
	BitGrade7
	BitGrade8
	BitGrade9
	BitGrade10
	BitGrade11
	BitGrade12
)

func (m GradesBitmask) IsGradeSet(b Grade) bool {
	if int(b) < 0 {
		return false
	}
	return int(m)&(1<<int(b)) != 0
}

func (m *GradesBitmask) ToggleGrade(b Grade) *GradesBitmask {
	// unassigned (-1) has no bit – ignore it.
	if int(b) < 0 {
		return m
	}
	*m ^= 1 << int(b) // bit 0 for K, bit 1 for 1, etc.
	return m
}

func (m GradesBitmask) String() string {
	return strconv.Itoa(int(m))
}

func (m GradesBitmask) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(m))
}

func (m *GradesBitmask) UnmarshalJSON(b []byte) error {
	// Try to unmarshal as int first
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*m = GradesBitmask(i)
		return nil
	}
	// fallback: try as string (e.g., "3")
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if i, err := strconv.Atoi(s); err == nil {
		*m = GradesBitmask(i)
		return nil
	}
	return fmt.Errorf("invalid grades bitmask: %s", string(b))
}
