package handlers

import (
	"bytes"
	"encoding/json"
	"insight-engine-backend/models"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecuteAdHocQuery(t *testing.T) {
	// Setup
	db := SetupTestDB()
	app := fiber.New()
	mockExecutor := new(MockQueryExecutor)

	// Create handler with mock executor and nil cache (for simplicity)
	handler := NewQueryHandler(mockExecutor, nil)

	// Register Route
	app.Post("/api/queries/execute", handler.ExecuteAdHocQuery)

	// Create a test connection
	conn := models.Connection{
		ID:   "test-conn-1",
		Name: "Test DB",
		Type: "postgres",
	}
	db.Create(&conn)

	t.Run("Valid Query", func(t *testing.T) {
		// Mock expectation
		mockResult := &models.QueryResult{
			Columns:  []string{"id", "name"},
			Rows:     [][]interface{}{{1, "Test"}},
			RowCount: 1,
		}

		// Expect Execute to be called.
		// Note: arguments matching might need tweaks if ctx or pointer addresses differ.
		// We use mock.Anything for context and connection pointer to be safe for now,
		// but we verify connection ID inside the handler logic implicitly by it being fetched.
		mockExecutor.On("Execute", mock.Anything, mock.MatchedBy(func(c *models.Connection) bool {
			return c.ID == "test-conn-1"
		}), "SELECT * FROM users", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

		payload := map[string]string{
			"connectionId": "test-conn-1",
			"sql":          "SELECT * FROM users",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/queries/execute", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		// Verify response body
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, true, result["success"])
	})

	t.Run("Missing Connection ID", func(t *testing.T) {
		payload := map[string]string{
			"sql": "SELECT * FROM users",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/queries/execute", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Connection Not Found", func(t *testing.T) {
		payload := map[string]string{
			"connectionId": "non-existent-conn",
			"sql":          "SELECT * FROM users",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/queries/execute", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	// Add more test cases as needed...
}
