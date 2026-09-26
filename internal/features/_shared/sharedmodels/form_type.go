package sharedmodels

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type FormType int

const (
	FormTypeUnknown FormType = iota
	FormTypeCreate
	FormTypeEdit
)

func (ft FormType) Word() string {
	switch ft {
	case FormTypeCreate:
		return "create"
	case FormTypeEdit:
		return "edit"
	default:
		return "unknown"
	}
}

func (ft FormType) String() string {
	return strconv.Itoa(int(ft))
}

func (ft FormType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Word())
}

func (ft *FormType) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*ft = FormType(int(val))
		return nil
	case string:
		switch val {
		case "":
			*ft = FormTypeUnknown
			return nil
		case "create":
			*ft = FormTypeCreate
			return nil
		case "edit":
			*ft = FormTypeEdit
			return nil
		case "unknown":
			*ft = FormTypeUnknown
			return nil
		}
		i, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("form type: unknown word or number %q", val)
		}
		*ft = FormType(i)
		return nil
	default:
		return fmt.Errorf("form type: unsupported type %T", v)
	}
}
