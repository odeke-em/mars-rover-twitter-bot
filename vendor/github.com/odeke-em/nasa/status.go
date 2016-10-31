package nasa

import (
	"encoding/json"
	"strings"
)

type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

func (s Status) String() string {
	return string(s)
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch strings.ToLower(str) {
	default:
		*s = Inactive
	case "active":
		*s = Active
	}
	return nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}
