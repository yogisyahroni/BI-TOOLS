package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const baseURL = "http://localhost:8080/api"

func registerUser(b *testing.B, email, password, username string) {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"username": username,
	})
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		b.Fatalf("Registration failed: %v", err)
	}
	defer resp.Body.Close()
	// Status code check skipped to allow existing users
}

func verifyUserInDB_Benchmark(b *testing.B, email string) {
	dsn := "postgresql://postgres:1234@localhost:5432/Inside_engineer1?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		b.Fatalf("Failed to connect to database for verification: %v", err)
	}

	if err := db.Exec("UPDATE users SET email_verified = true, email_verified_at = NOW(), status = 'active' WHERE email = ?", email).Error; err != nil {
		b.Fatalf("Failed to verify user in DB: %v", err)
	}
}

func loginUser(b *testing.B, email, password string) string {
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		b.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b.Fatalf("Login failed status: %d", resp.StatusCode)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		b.Fatalf("Login decode failed: %v", err)
	}

	// Handle token location variations if necessary
	if token, ok := res["token"].(string); ok {
		return token
	}
	if data, ok := res["data"].(map[string]interface{}); ok {
		if t, ok := data["token"].(string); ok {
			return t
		}
	}
	b.Fatalf("Token not found in login response: %v", res)
	return ""
}

func createConnection(b *testing.B, token string) string {
	conn := map[string]interface{}{
		"name":     "Benchmark DB " + uuid.New().String(),
		"type":     "postgres",
		"host":     "localhost",
		"port":     5432,
		"database": "Inside_engineer1",
		"username": "postgres",
		"password": "1234",
	}
	body, _ := json.Marshal(conn)
	req, _ := http.NewRequest("POST", baseURL+"/connections", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Fatalf("Create connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b.Fatalf("Create connection failed status: %d", resp.StatusCode)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		b.Fatalf("Connection data not found: %v", res)
	}
	return data["id"].(string)
}

func createQuery(b *testing.B, token, connID string) string {
	query := map[string]interface{}{
		"name":         "Benchmark Query " + uuid.New().String(),
		"sql":          "SELECT 1",
		"connectionId": connID,
	}
	body, _ := json.Marshal(query)
	req, _ := http.NewRequest("POST", baseURL+"/queries", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.Fatalf("Create query failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b.Fatalf("Create query failed status: %d", resp.StatusCode)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		b.Fatalf("Query data not found: %v", res)
	}
	return data["id"].(string)
}

func BenchmarkQueryExecution(b *testing.B) {
	uniqueID := uuid.New().String()
	email := fmt.Sprintf("bench_%s@example.com", uniqueID)
	password := "Test@1234"
	username := fmt.Sprintf("BenchUser_%s", uniqueID)

	registerUser(b, email, password, username)
	verifyUserInDB_Benchmark(b, email)
	token := loginUser(b, email, password)
	connID := createConnection(b, token)
	queryID := createQuery(b, token, connID)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", baseURL+"/queries/"+queryID+"/run", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("Run request failed: %v", err)
		}

		if resp.StatusCode != 200 {
			b.Errorf("Query run failed: %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	type Row struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Value     float64     `json:"value"`
		Timestamp string      `json:"timestamp"`
		Meta      interface{} `json:"meta"`
	}

	// Create 1000 rows
	rows := make([]Row, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = Row{
			ID:        uuid.New().String(),
			Name:      fmt.Sprintf("Item %d", i),
			Value:     float64(i) * 1.5,
			Timestamp: time.Now().Format(time.RFC3339),
			Meta:      map[string]interface{}{"active": true, "score": 99},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(rows)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
	}
}
