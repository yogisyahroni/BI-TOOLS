package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestQueryFlow(t *testing.T) {
	// 1. Setup User
	uniqueID := uuid.New().String()
	email := fmt.Sprintf("query_test_%s@example.com", uniqueID)
	password := "Test@1234"
	username := fmt.Sprintf("QueryUser_%s", uniqueID)

	registerUser(t, email, password, username)
	// Verify user to allow login (auto-verification might not be enabled)
	verifyUserInDB(t, email)
	token := loginUser(t, email, password)

	// 2. Create Connection
	var connectionID string
	t.Run("Create Info Connection", func(t *testing.T) {
		conn := map[string]interface{}{
			"name":     "Test DB Connection",
			"type":     "postgres",
			"host":     "localhost",
			"port":     5432,
			"database": "Inside_engineer1",
			"username": "postgres",
			"password": "1234", // Using the known credential
		}
		body, _ := json.Marshal(conn)
		req, _ := http.NewRequest("POST", baseURL+"/connections", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		data := res["data"].(map[string]interface{})
		connectionID = data["id"].(string)
	})

	if connectionID == "" {
		t.Fatal("Connection ID is empty, skipping remaining tests")
	}

	// 3. Create Saved Query
	var queryID string
	t.Run("Create Saved Query", func(t *testing.T) {
		query := map[string]interface{}{
			"name":         "Test Query",
			"sql":          "SELECT 1 as result",
			"connectionId": connectionID,
		}
		body, _ := json.Marshal(query)
		req, _ := http.NewRequest("POST", baseURL+"/queries", bytes.NewBuffer(body)) // Assuming /queries based on setup.go
		// Wait, setup.go says `api.Post("/queries", ...)`
		// But query_handler.go comments said `/query`.
		// Let's try `/query` if `/queries` fails, or check setup.go again.
		// setup.go: `api.Post("/queries", m.AuthMiddleware, h.QueryHandler.CreateQuery)`
		// So `/api/queries` is correct. HOWEVER, checking lines 49-55 of setup.go:
		/*
			api.Get("/queries", m.AuthMiddleware, h.QueryHandler.GetQueries)
			api.Post("/queries", m.AuthMiddleware, h.QueryHandler.CreateQuery)
		*/
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Create Query failed: %s", string(bodyBytes))
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		// setup.go CreateQuery returns `c.Status(201).JSON(fiber.Map{"success": true, "data": query})`
		data := res["data"].(map[string]interface{})
		queryID = data["id"].(string)
	})

	// 4. Run Saved Query
	t.Run("Run Saved Query", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/queries/"+queryID+"/run", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Run Query failed: %s", string(bodyBytes))
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		// data might be result rows?
		// result format depends on `queryExecutor`.
	})

	// 5. Run Ad-Hoc Query
	t.Run("Run Ad-Hoc Query", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"sql":          "SELECT 'hello' as greeting",
			"connectionId": connectionID,
		})
		req, _ := http.NewRequest("POST", baseURL+"/queries/execute", bytes.NewBuffer(body)) // setup.go line 55
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 6. Delete Query
	t.Run("Delete Query", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", baseURL+"/queries/"+queryID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
