package convert

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

func DecodeJSON[T any](raw datatypes.JSON) (T, error) {
	var out T
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode json: %w", err)
	}
	return out, nil
}

func EncodeJSON(v any) (datatypes.JSON, error) {
	if v == nil {
		return datatypes.JSON([]byte("null")), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}
	return datatypes.JSON(b), nil
}
