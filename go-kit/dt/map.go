package dt

import (
	"database/sql/driver"
	"encoding/json"
)

type Map map[string]interface{}

func (s Map) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Map) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return json.Unmarshal(value.([]byte), s)
	}
}

func (s *Map) Add(key string, value any) {
	if *s == nil {
		*s = make(Map)
	}

	(*s)[key] = value
}

func (s Map) Contains(key string) bool {
	_, found := s[key]
	return found
}
