package redis

import (
	"encoding/json"
	"time"
)

const NoTTL time.Duration = 0

type JSONObject[T any] struct {
	Data T `json:"data"`
}

func (g JSONObject[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(g)
}

func (g *JSONObject[T]) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, g)
}
