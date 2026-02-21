package handlers

import (
	"encoding/json"
	"insight-engine-backend/models"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConnections_Success(t *testing.T) {
	db := SetupTestDB()
	app := fiber.New()
	handler := NewConnectionHandler(nil, nil, nil)

	app.Get("/connection", func(c *fiber.Ctx) error {
		c.Locals("userId", "user-123") // Mock Auth Middleware
		return handler.GetConnections(c)
	})

	// Seed data
	db.Create(&models.Connection{
		ID:     "conn-1",
		UserID: "user-123",
		Name:   "Prod DB",
		Type:   "postgres",
	})
	db.Create(&models.Connection{
		ID:     "conn-2",
		UserID: "user-123",
		Name:   "Staging DB",
		Type:   "mysql",
	})

	req := httptest.NewRequest("GET", "/connection", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, true, result["success"])

	data := result["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestGetConnection_Success(t *testing.T) {
	db := SetupTestDB()
	app := fiber.New()
	handler := NewConnectionHandler(nil, nil, nil)

	app.Get("/connection/:id", func(c *fiber.Ctx) error {
		c.Locals("userId", "user-123")
		return handler.GetConnection(c)
	})

	db.Create(&models.Connection{
		ID:     "conn-1",
		UserID: "user-123",
		Name:   "Prod DB",
	})

	req := httptest.NewRequest("GET", "/connection/conn-1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetConnection_NotFound(t *testing.T) {
	SetupTestDB() // Ensure fresh DB
	app := fiber.New()
	handler := NewConnectionHandler(nil, nil, nil)

	app.Get("/connection/:id", func(c *fiber.Ctx) error {
		c.Locals("userId", "user-123")
		return handler.GetConnection(c)
	})

	req := httptest.NewRequest("GET", "/connection/non-existent", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestTestConnection_Success(t *testing.T) {
	db := SetupTestDB()
	app := fiber.New()
	mockExecutor := new(MockQueryExecutor)

	// Create handler with mock executor
	handler := NewConnectionHandler(mockExecutor, nil, nil)

	app.Post("/connection/:id/test", func(c *fiber.Ctx) error {
		c.Locals("userId", "user-123")
		return handler.TestConnection(c)
	})

	db.Create(&models.Connection{
		ID:     "conn-test-1",
		UserID: "user-123",
		Name:   "Prod DB",
		Type:   "postgres",
	})

	// Expect Execute to be called with SELECT 1
	mockExecutor.On("Execute", mock.Anything, mock.Anything, "SELECT 1", mock.Anything, mock.Anything, mock.Anything).Return(&models.QueryResult{}, nil)

	req := httptest.NewRequest("POST", "/connection/conn-test-1/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "success", result["status"])
}

func TestTestConnection_MockBackdoor(t *testing.T) {
	db := SetupTestDB()
	app := fiber.New()
	mockExecutor := new(MockQueryExecutor)
	handler := NewConnectionHandler(mockExecutor, nil, nil)

	app.Post("/connection/:id/test", func(c *fiber.Ctx) error {
		c.Locals("userId", "user-123")
		return handler.TestConnection(c)
	})

	db.Create(&models.Connection{
		ID:     "conn-mock-1",
		UserID: "user-123",
		Name:   "TestDB-Mock", // Starts with TestDB-
		Type:   "postgres",
	})

	// Execute should NOT be called
	req := httptest.NewRequest("POST", "/connection/conn-mock-1/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "Connection successful (MOCKED)", result["message"])
}
