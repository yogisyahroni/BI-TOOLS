package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const baseURL = "http://localhost:8080/api"
const dbDSN = "postgresql://postgres:1234@localhost:5432/Inside_engineer1?sslmode=disable"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type SemanticModel struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	DataSourceID string              `json:"dataSourceId"` // Use a dummy ID or fetch
	TableName    string              `json:"tableName"`
	Dimensions   []SemanticDimension `json:"dimensions"`
	Metrics      []SemanticMetric    `json:"metrics"`
}

type SemanticDimension struct {
	Name       string `json:"name"`
	ColumnName string `json:"columnName"`
	DataType   string `json:"dataType"`
	IsHidden   bool   `json:"isHidden"`
}

type SemanticMetric struct {
	Name    string `json:"name"`
	Formula string `json:"formula"`
	Format  string `json:"format"`
}

// Minimal User struct for DB query
type User struct {
	EmailVerificationToken string
}

type ConnectionRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Database string `json:"database"`
}

type ConnectionResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID string `json:"id"`
	} `json:"data"`
}

type WorkspaceRequest struct {
	Name string `json:"name"`
}

type WorkspaceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func TestSemanticLayerCRUD(t *testing.T) {
	// 0. Register unique user
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email := "test_user_" + uniqueSuffix + "@example.com"
	password := "password123"

	register(t, email, password, uniqueSuffix)

	// 0.5 Verify Email (Bypass via DB)
	verifyEmail(t, email)

	// 1. Login
	token := login(t, email, password)

	// 1.2 Create Workspace
	workspaceID := createWorkspace(t, token)

	// 1.5 Create Connection (DataSource)
	connectionID := createConnection(t, token, workspaceID)

	// 2. Create Model
	model := createModel(t, token, workspaceID, connectionID)

	// 3. Get Model
	fetchedModel := getModel(t, token, workspaceID, model.ID)
	if fetchedModel.Name != model.Name {
		t.Errorf("Expected model name %s, got %s", model.Name, fetchedModel.Name)
	}

	// 4. Update Model (Add Metric)
	fetchedModel.Metrics = append(fetchedModel.Metrics, SemanticMetric{
		Name:    "NewMetric",
		Formula: "SUM(price)",
		Format:  "currency",
	})
	updateModel(t, token, workspaceID, fetchedModel)

	// Verify Update
	updatedModel := getModel(t, token, workspaceID, model.ID)
	if len(updatedModel.Metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(updatedModel.Metrics))
	}

	// 5. Delete Model
	deleteModel(t, token, workspaceID, model.ID)

	// Verify Deletion
	verifyDeletion(t, token, workspaceID, model.ID)
}

func register(t *testing.T, email, password, suffix string) {
	reqBody, _ := json.Marshal(RegisterRequest{
		Email:    email,
		Password: password,
		Name:     "Test User " + suffix,
		Username: "user_" + suffix,
	})
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Register failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func verifyEmail(t *testing.T, email string) {
	// Connect to DB to get token
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database for verification: %v", err)
	}

	var user User
	// Query users table. Note: Table name is usually plural 'users'
	if err := db.Table("users").Select("email_verification_token").Where("email = ?", email).First(&user).Error; err != nil {
		t.Fatalf("Failed to find user in database: %v", err)
	}

	if user.EmailVerificationToken == "" {
		t.Fatalf("User has no verification token")
	}

	// Call Verification Endpoint
	reqURL := fmt.Sprintf("%s/auth/verify-email?token=%s", baseURL, user.EmailVerificationToken)
	resp, err := http.Get(reqURL)
	if err != nil {
		t.Fatalf("Verification request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Verification failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func login(t *testing.T, email, password string) string {
	reqBody, _ := json.Marshal(LoginRequest{
		Email:    email,
		Password: password,
	})
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	return loginResp.Token
}

func createWorkspace(t *testing.T, token string) string {
	wsReq := WorkspaceRequest{
		Name: "Test Workspace " + generateRandomString(),
	}
	reqBody, _ := json.Marshal(wsReq)
	req, _ := http.NewRequest("POST", baseURL+"/workspaces", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Create Workspace failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Create Workspace failed with status %d: %s", resp.StatusCode, string(body))
	}

	var wsResp WorkspaceResponse
	json.NewDecoder(resp.Body).Decode(&wsResp)
	return wsResp.ID
}

func createConnection(t *testing.T, token string, workspaceID string) string {
	conn := ConnectionRequest{
		Name:     "Test Connection " + generateRandomString(),
		Type:     "postgres",
		Database: "test_db",
	}
	reqBody, _ := json.Marshal(conn)
	req, _ := http.NewRequest("POST", baseURL+"/connections", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Create Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Create Connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	var connResp ConnectionResponse
	json.NewDecoder(resp.Body).Decode(&connResp)
	return connResp.Data.ID
}

func createModel(t *testing.T, token string, workspaceID string, dataSourceID string) SemanticModel {
	model := SemanticModel{
		Name:         "Test Model " + generateRandomString(),
		Description:  "Integration Test Model",
		DataSourceID: dataSourceID,
		TableName:    "public.test_table",
		Dimensions:   []SemanticDimension{},
		Metrics:      []SemanticMetric{},
	}

	reqBody, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", baseURL+"/semantic/models", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Create Model failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Create Model failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createdModel SemanticModel
	json.Unmarshal(body, &createdModel)
	return createdModel
}

func getModel(t *testing.T, token string, workspaceID string, id string) SemanticModel {
	req, _ := http.NewRequest("GET", baseURL+"/semantic/models/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Get Model failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get Model failed with status %d", resp.StatusCode)
	}

	var model SemanticModel
	json.NewDecoder(resp.Body).Decode(&model)
	return model
}

func updateModel(t *testing.T, token string, workspaceID string, model SemanticModel) {
	reqBody, _ := json.Marshal(model)
	req, _ := http.NewRequest("PUT", baseURL+"/semantic/models/"+model.ID, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Update Model failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Update Model failed with status %d: %s", resp.StatusCode, string(body))
	}
}

func deleteModel(t *testing.T, token string, workspaceID string, id string) {
	req, _ := http.NewRequest("DELETE", baseURL+"/semantic/models/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Delete Model failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Delete Model failed with status %d", resp.StatusCode)
	}
}

func verifyDeletion(t *testing.T, token string, workspaceID string, id string) {
	req, _ := http.NewRequest("GET", baseURL+"/semantic/models/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Workspace-ID", workspaceID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Verify Deletion failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404 for deleted model, got %d", resp.StatusCode)
	}
}

func generateRandomString() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
