package datatypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap represents a map that is stored as JSON in the database
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSONB value: invalid type")
	}

	result := make(map[string]interface{})
	err := json.Unmarshal(bytes, &result)
	*m = JSONMap(result)
	return err
}
