package tests

import (
	"encoding/json"
	"insight-engine-backend/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccelerationService_Execution(t *testing.T) {
	// Get Singleton
	accel := services.GetAccelerationService()
	assert.NotNil(t, accel)

	// Test Simple Query (SQLite syntax)
	result, err := accel.ExecuteQuery("SELECT 1 AS id, 'sqlite' AS name")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Rows))
	assert.Equal(t, "id", result.Columns[0])
	assert.Equal(t, "name", result.Columns[1])

	// Value check
	row1 := result.Rows[0]
	// SQLite driver typically returns int64 for integers
	assert.Equal(t, int64(1), row1[0])
	assert.Equal(t, "sqlite", row1[1])
}

func TestAccelerationService_LoadJSON(t *testing.T) {
	accel := services.GetAccelerationService()

	// Prepare Dummy Data using map[string]interface{}
	jsonData := `[
		{"id": 1, "name": "Alice", "score": 95.5, "active": true},
		{"id": 2, "name": "Bob", "score": 80.0, "active": false}
	]`
	var data []interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	assert.NoError(t, err)

	// Load Data
	err = accel.LoadJSON("users_test", data)
	assert.NoError(t, err)

	// Query Data
	result, err := accel.ExecuteQuery("SELECT * FROM users_test ORDER BY id")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Rows))

	row1 := result.Rows[0]
	// Order depends on map iteration during creation, so columns might be shuffled
	// Let's find index of "name"
	nameIdx := -1
	for i, col := range result.Columns {
		if col == "name" {
			nameIdx = i
			break
		}
	}
	assert.True(t, nameIdx >= 0)
	assert.Equal(t, "Alice", row1[nameIdx])
}
